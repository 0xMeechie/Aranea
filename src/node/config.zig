const std = @import("std");
const fs = std.fs;
const json = std.json;
const id = @import("./identity.zig");
const hash = std.crypto.hash;
const logly = @import("logly");

pub const configError = error{HomeNotSet};

const DEFAULT_CONFIG_LOCATION = "/.config/thermosphere";
const CONFIG_FILE = "config.json";
const IDENTITY_FILE = "identity.json";
const AUDIT_FILE = "audit-000001.ael";
const CONFIG_VERSION = 1;
const DEFAULT_FILE_PERMISSIONS = 0o600;
const DEFAULT_FOLDER_PERMISSIONS = 0o700;

const logLevels = enum { info, debug, verbose };

const identityData = struct {
    version: u4 = CONFIG_VERSION,
    public_key: []const u8,
    private_key: []const u8,
    created_at: i64,
    node_name: []const u8,
};
const configData = struct {
    version: u4 = CONFIG_VERSION,
    config_dir: []const u8,
    config_file: []const u8,
    identity_file: []const u8,
    audit_dir: []const u8,
    log_level: logLevels,
};

// private struct to remove the possibility to overwrite the config path
const configPath = struct {
    path: []const u8,
};

pub const config = struct {
    allocator: std.mem.Allocator,
    configPath: configPath,
    version: u4 = CONFIG_VERSION,
    logger: *logly.Logger,

    pub fn currentConfigPath(self: *config) []const u8 {
        return self.configPath.path;
    }

    pub fn init(allocator: std.mem.Allocator, logger: *logly.Logger) !config {
        const homeDir = std.process.getEnvVarOwned(allocator, "HOME") catch return configError.HomeNotSet;
        defer allocator.free(homeDir);
        const configDir = try std.mem.concat(allocator, u8, &.{ homeDir, DEFAULT_CONFIG_LOCATION });

        return .{ .configPath = configPath{ .path = configDir }, .allocator = allocator, .logger = logger };
    }

    fn createThermoDirectories(self: *config) !void {
        const idDir = try std.mem.concat(self.allocator, u8, &.{ self.currentConfigPath(), "/identity" });
        const configDir = try std.mem.concat(self.allocator, u8, &.{ self.currentConfigPath(), "/config" });
        const auditDir = try std.mem.concat(self.allocator, u8, &.{ self.currentConfigPath(), "/audit" });
        defer self.allocator.free(idDir);
        defer self.allocator.free(configDir);
        defer self.allocator.free(auditDir);
        const dirs = [_][]const u8{
            idDir,
            configDir,
            auditDir,
        };

        for (dirs) |dir| {
            fs.cwd().makeDir(dir) catch |e| {
                if (e == fs.Dir.MakeError.PathAlreadyExists) {
                    try self.logger.debugf("{s} exist already. Skipping Creation \n", .{dir}, @src());
                }
                continue;
            };
            var openDir = try fs.cwd().openDir(dir, .{});
            openDir.chmod(DEFAULT_FOLDER_PERMISSIONS) catch {
                try self.logger.debugf("unable to update permissions for {s}. Continuing Setup.. \n", .{dir}, @src());
            };

            openDir.close();
            try self.logger.debugf("Successfully created directory [{s}] with the correct permissions \n", .{dir}, @src());
        }
    }

    pub fn configPathsExist(self: *config) bool {
        const homeDir = std.process.getEnvVarOwned(self.allocator, "HOME") catch return configError.HomeNotSet;
        defer self.allocator.free(homeDir);
    }

    pub fn initNodeConfig(self: *config, forceInit: bool) !void {
        try self.createThermoDirectories();
        const keys = try id.generateIdentity(self.allocator);
        defer self.allocator.free(keys.public_key);
        defer self.allocator.free(keys.private_key);
        try self.createConfigFile(forceInit);
        try self.createIdentityFile(keys, forceInit);
        try self.createAuditFile();
    }

    fn createAuditFile(self: *config) !void {
        const filePath = try self.auditFilePath();
        defer self.allocator.free(filePath);

        var fileExist = true;

        //check to see if the file already exist
        fs.cwd().access(filePath, .{}) catch |e| {
            if (e == fs.Dir.AccessError.FileNotFound) {
                try self.logger.debugf("{s} does not exist in the following directory {s}. Attempting to create now \n", .{ AUDIT_FILE, self.currentConfigPath() }, @src());
                fileExist = false;
            } else {
                return e;
            }
        };

        // never want to overwrite the audit files
        if (fileExist) {
            try self.logger.debugf("{s} already exist in the following directory {s}.\n", .{ AUDIT_FILE, self.currentConfigPath() }, @src());
            return;
        }
        var file = try fs.createFileAbsolute(filePath, .{});
        file.chmod(DEFAULT_FILE_PERMISSIONS) catch {
            try self.logger.debugf("[WARNING] {s} couldn't be created with the correct permissions. Continuing Setup..", .{filePath}, @src());
        };

        try self.logger.debugf("successfully created {s} in the following directory {s}.\n", .{ AUDIT_FILE, self.currentConfigPath() }, @src());
    }

    fn createIdentityFile(self: *config, keys: id.nodeIdentity, forcedOverwrite: bool) !void {
        const filePath = try self.identityFilePath();
        defer self.allocator.free(filePath);

        var fileExist = true;

        //check to see if the file already exist
        fs.cwd().access(filePath, .{}) catch |e| {
            if (e == fs.Dir.AccessError.FileNotFound) {
                try self.logger.debugf("{s} does not exist in the following directory {s}. Attempting to create now \n", .{ IDENTITY_FILE, self.currentConfigPath() }, @src());
                fileExist = false;
            } else {
                return e;
            }
        };

        if (fileExist and !forcedOverwrite) {
            try self.logger.debugf("{s} already exist in the following directory {s}.\n", .{ IDENTITY_FILE, self.currentConfigPath() }, @src());
            return;
        }

        try self.logger.debugf("successfully created {s} in the following directory {s}.\n", .{ IDENTITY_FILE, self.currentConfigPath() }, @src());
        var file = try fs.createFileAbsolute(filePath, .{});
        file.chmod(DEFAULT_FILE_PERMISSIONS) catch {
            try self.logger.warningf("{s} couldn't be created with the correct permissions. Continuing Setup..", .{filePath}, @src());
        };
        defer file.close();
        var full_hash: [32]u8 = undefined;
        hash.sha2.Sha256.hash(keys.public_key, &full_hash, .{});

        const hex = std.fmt.bytesToHex(full_hash, .lower);
        const idData = identityData{
            .public_key = keys.public_key,
            .private_key = keys.private_key,
            .node_name = &hex,
            .created_at = std.time.timestamp(),
            .version = CONFIG_VERSION,
        };
        var out = std.Io.Writer.Allocating.init(self.allocator);
        const writer = &out.writer;
        defer out.deinit();
        try json.Stringify.value(idData, .{ .whitespace = .indent_tab }, writer);

        const writtenData = out.written();

        try file.writeAll(writtenData);
        try writer.flush();
    }

    fn configFilePath(self: *config) ![]u8 {
        const filePath = try std.mem.concat(self.allocator, u8, &.{ self.configPath.path, "/config/", CONFIG_FILE });
        return filePath;
    }

    fn identityFilePath(self: *config) ![]u8 {
        const filePath = try std.mem.concat(self.allocator, u8, &.{ self.configPath.path, "/identity/", IDENTITY_FILE });
        return filePath;
    }

    fn auditFilePath(self: *config) ![]u8 {
        const filePath = try std.mem.concat(self.allocator, u8, &.{ self.configPath.path, "/audit/", AUDIT_FILE });
        return filePath;
    }

    fn createConfigFile(self: *config, forceOverwrite: bool) !void {
        const filePath = try self.configFilePath();
        defer self.allocator.free(filePath);

        var fileExist = true;

        fs.cwd().access(filePath, .{}) catch |e| {
            if (e == fs.Dir.AccessError.FileNotFound) {
                try self.logger.debugf("{s} does not exist in the following directory {s}. Attempting to create now \n", .{ CONFIG_FILE, self.currentConfigPath() }, @src());
                fileExist = false;
            } else {
                return e;
            }
        };

        if (fileExist and !forceOverwrite) {
            try self.logger.debugf("{s} already exist in the following directory {s}.\n", .{ CONFIG_FILE, self.currentConfigPath() }, @src());
            return;
        }

        var file = try fs.createFileAbsolute(filePath, .{});
        try file.chmod(DEFAULT_FILE_PERMISSIONS);

        file.chmod(DEFAULT_FILE_PERMISSIONS) catch {
            try self.logger.debugf("[WARNING] {s} couldn't be created with the correct permissions. Continuing Setup..", .{filePath}, @src());
        };
        defer file.close();
        const idPath = try self.identityFilePath();
        defer self.allocator.free(idPath);
        const cfgFilePath = try self.configFilePath();
        defer self.allocator.free(cfgFilePath);
        const auditDir = try std.mem.concat(self.allocator, u8, &.{ self.currentConfigPath(), "/audit" });
        defer self.allocator.free(auditDir);
        const cfdDir = self.currentConfigPath();
        const cfgData = configData{ .config_dir = cfdDir, .identity_file = idPath, .version = CONFIG_VERSION, .log_level = logLevels.info, .config_file = cfgFilePath, .audit_dir = auditDir };

        var out = std.Io.Writer.Allocating.init(self.allocator);
        const writer = &out.writer;
        defer out.deinit();
        try json.Stringify.value(cfgData, .{ .whitespace = .indent_tab }, writer);

        const writtenData = out.written();

        try file.writeAll(writtenData);
        try writer.flush();

        try self.logger.debugf("successfully created {s} in the following directory {s}.\n", .{ CONFIG_FILE, self.currentConfigPath() }, @src());
    }

    pub fn deinit(self: *config) void {
        self.allocator.free(self.configPath.path);
    }
};
