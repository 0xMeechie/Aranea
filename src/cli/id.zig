const std = @import("std");
const Writer = std.Io.Writer;
const Reader = std.Io.Reader;
const zli = @import("zli");
const logly = @import("logly");
const config = @import("config");

pub fn register(writer: *Writer, reader: *Reader, allocator: std.mem.Allocator) !*zli.Command {
    return zli.Command.init(writer, reader, allocator, .{
        .name = "id",
        .shortcut = "id",
        .description = "return the id of the node",
    }, showID);
}

fn showID(ctx: zli.CommandContext) !void {
    var logger = try logly.Logger.init(ctx.allocator);
    var lConfig = logly.Config.default();
    lConfig.level = .debug;
    defer logger.deinit();
    logger.configure(lConfig);
    var nc = try config.config.init(ctx.allocator, logger);
    defer nc.deinit();

    const valid = try nc.isValid();

    if (!valid) {
        try logger.err("this node is not config correctly. Please try to reinitialize the node again.", @src());
        return;
    }

    const nodeName = try nc.nodeNamme();
    defer ctx.allocator.free(nodeName);

    std.debug.print("Node Name: {s} \n ", .{nodeName});
}
