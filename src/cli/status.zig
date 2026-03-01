const std = @import("std");
const Writer = std.Io.Writer;
const Reader = std.Io.Reader;
const zli = @import("zli");
const logly = @import("logly");

const config = @import("config");

pub fn register(writer: *Writer, reader: *Reader, allocator: std.mem.Allocator) !*zli.Command {
    const cmd = try zli.Command.init(writer, reader, allocator, .{
        .name = "status",
        .shortcut = "s",
        .description = "returns the status of the current node",
    }, showStatus);

    return cmd;
}

fn showStatus(ctx: zli.CommandContext) !void {
    // check to see if the config and identity file exist
    // if it doesn't create it.
    // if it does there's nothing to do
    // if force is enabled then rewrite the config and identity files
    var logger = try logly.Logger.init(ctx.allocator);
    var lConfig = logly.Config.default();
    lConfig.level = .debug;
    defer logger.deinit();
    logger.configure(lConfig);
    try logger.info("checking status of the node", @src());
    var nc = try config.config.init(ctx.allocator, logger);
    defer nc.deinit();

    // if forcedEnabled we will rewrite the current configs and Init a new node config. If disabled we will only create
    // the initial node config. If we init again ignore any work that needs to be done.

    try logger.info("NODE INFO \n", @src());
}
