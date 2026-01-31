const std = @import("std");

// generate ID for the node.

pub const nodeIdentity = struct { public_key: []const u8, private_key: []const u8 };
pub fn generateIdentity(allocator: std.mem.Allocator) !nodeIdentity {
    const kp = std.crypto.sign.Ed25519.KeyPair.generate();
    const pubKey = kp.public_key.bytes;
    const priKey = kp.secret_key.bytes;

    const prik = try allocator.alloc(u8, std.base64.standard.Encoder.calcSize(priKey.len));
    const pubk = try allocator.alloc(u8, std.base64.standard.Encoder.calcSize(pubKey.len));

    const k = std.base64.standard.Encoder.encode(pubk, &pubKey);

    const pk = std.base64.standard.Encoder.encode(prik, &priKey);

    return .{ .public_key = k, .private_key = pk };
}
