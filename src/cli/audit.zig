const std = @import("std");
const Writer = std.Io.Writer;
const Reader = std.Io.Reader;
const zli = @import("zli");
const logly = @import("logly");

pub fn register(writer: *Writer, reader: *Reader, allocator: std.mem.Allocator) !*zli.Command {
    return zli.Command.init(writer, reader, allocator, .{
        .name = "audit",
        .shortcut = "a",
        .description = "view recent events",
    }, showID);
}

fn showID(ctx: zli.CommandContext) !void {
    var logger = try logly.Logger.init(ctx.allocator);
    var lConfig = logly.Config.default();
    lConfig.level = .debug;
    defer logger.deinit();
    logger.configure(lConfig);
    try logger.info("starting audit of events", @src());
}
