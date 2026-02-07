const std = @import("std");
const Writer = std.Io.Writer;
const Reader = std.Io.Reader;
const zli = @import("zli");

const version = @import("version.zig");
const ping = @import("ping.zig");
const id = @import("id.zig");
const init = @import("init.zig");
const serve = @import("serve.zig");
const status = @import("status.zig");

pub fn build(writer: *Writer, reader: *Reader, allocator: std.mem.Allocator) !*zli.Command {
    const root = try zli.Command.init(writer, reader, allocator, .{
        .name = "thermosphere",
        .description = "thermosphere is a cli tool for interacting with nodes",
        .version = .{ .major = 0, .minor = 0, .patch = 1, .pre = null, .build = null },
    }, showHelp);

    try root.addCommands(&.{
        try version.register(writer, reader, allocator),
        try id.register(writer, reader, allocator),
        try ping.register(writer, reader, allocator),
        try init.register(writer, reader, allocator),
        try serve.register(writer, reader, allocator),
        try status.register(writer, reader, allocator),
    });

    return root;
}

fn showHelp(ctx: zli.CommandContext) !void {
    try ctx.command.printHelp();
}
