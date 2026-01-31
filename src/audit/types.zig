const std = @import("std");

pub const AEL_MAGIC = "AEL0";
pub const AEL_VERSION: u16 = 0;
pub const AEL_HEADER_LEN: u16 = 64;

const SegmentHeader = packed struct {
    magic: [4]u8,
    version: u16,
    header_len: u16,
    flags: u32,
    created_at_unix_ms: i64,
    segment_id: u64,
    reserved: [32]u8, // zero-filled
    header_crc32: u32,
    name: u8,
};

comptime {
    // Hard safety check — never let this drift
    if (@sizeOf(SegmentHeader) != 64) {
        @compileError("SegmentHeader must be exactly 64 bytes");
    }
}
