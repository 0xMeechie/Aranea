const std = @import("std");
const fs = std.fs;
const cli = @import("cli");

pub fn main() !void {
    var dbg = std.heap.DebugAllocator(.{}).init;

    const allocator = switch (@import("builtin").mode) {
        .Debug => dbg.allocator(),
        .ReleaseFast, .ReleaseSafe, .ReleaseSmall => std.heap.smp_allocator,
    };

    defer if (@import("builtin").mode == .Debug) std.debug.assert(dbg.deinit() == .ok);

    var stdout_writer = fs.File.stdout().writerStreaming(&.{});
    const stdout = &stdout_writer.interface;

    var buf: [4096]u8 = undefined;
    var stdin_reader = fs.File.stdin().readerStreaming(&buf);
    const stdin = &stdin_reader.interface;

    const root = try cli.build(stdout, stdin, allocator);
    defer root.deinit();

    try root.execute(.{}); // Or pass data with: try root.execute(.{ .data = &my_data });

    try stdout_writer.interface.flush(); // Don't forget to flush!
}
