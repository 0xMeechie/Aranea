const std = @import("std");

pub fn build(b: *std.Build) void {
    // Standard target options allow the person running `zig build` to choose
    // what target to build for. Here we do not override the defaults, which
    // means any target is allowed, and the default is native. Other options
    // for restricting supported target set are available.
    const target = b.standardTargetOptions(.{});
    // Standard optimization options allow the person running `zig build` to select
    // between Debug, ReleaseSafe, ReleaseFast, and ReleaseSmall. Here we do not
    // set a preferred release mode, allowing the user to decide how to optimize.
    const optimize = b.standardOptimizeOption(.{});
    const zli_dep = b.dependency("zli", .{ .target = target, .optimize = optimize });
    const logly_dep = b.dependency("logly", .{
        .target = target,
        .optimize = optimize,
    });

    const arenead = b.addExecutable(.{
        .name = "arenead",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/cmd/aranead/main.zig"),
            .target = target,
            .optimize = optimize,
        }),
    });

    const arenea_gateway = b.addExecutable(.{
        .name = "arenea-gateway",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/cmd/gateway/main.zig"),
            .target = target,
            .optimize = optimize,
        }),
    });

    const areneacli = b.addExecutable(.{
        .name = "areneacli",
        .root_module = b.createModule(.{
            .root_source_file = b.path("src/cmd/araneacli/main.zig"),
            .target = target,
            .optimize = optimize,
        }),
    });

    const cli_module = b.createModule(.{
        .root_source_file = b.path("src/cli/root.zig"),
        .target = target,
        .optimize = optimize,
    });
    const config_module = b.createModule(.{
        .root_source_file = b.path("src/node/config.zig"),
        .target = target,
        .optimize = optimize,
    });

    config_module.addImport("logly", logly_dep.module("logly"));

    cli_module.addImport("zli", zli_dep.module("zli"));
    cli_module.addImport("logly", logly_dep.module("logly"));
    cli_module.addImport("config", config_module);
    areneacli.root_module.addImport("cli", cli_module);

    areneacli.root_module.addImport("config", config_module);
    areneacli.root_module.addImport("logly", logly_dep.module("logly"));

    areneacli.root_module.addImport("zli", zli_dep.module("zli"));

    b.installArtifact(areneacli);
    b.installArtifact(arenead);
    b.installArtifact(arenea_gateway);

    const run_step = b.step("run", "Run the app");

    const run_cmd = b.addRunArtifact(areneacli);
    run_step.dependOn(&run_cmd.step);

    run_cmd.step.dependOn(b.getInstallStep());

    if (b.args) |args| {
        run_cmd.addArgs(args);
    }
}
