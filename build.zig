const std = @import("std");
const microzig = @import("microzig");

const MicroBuild = microzig.MicroBuild(.{
    .rp2xxx = true,
});

pub fn build(b: *std.Build) void {
    const mz_dep = b.dependency("microzig", .{});
    const mb = MicroBuild.init(b, mz_dep) orelse return;
    const target = mb.ports.rp2xxx.boards.raspberrypi.pico;
    const optimize = b.standardOptimizeOption(.{});

    const ir_dep = b.dependency("ir", .{});

    const firmware = mb.add_firmware(
        .{
            .name = "amp",
            .target = target,
            .optimize = optimize,
            .root_source_file = b.path("src/main.zig"),
            .imports = &.{
                .{ .name = "ir", .module = ir_dep.module("ir") },
            },
        },
    );

    const test_firmware = mb.add_firmware(
        .{
            .name = "blink",
            .target = target,
            .optimize = optimize,
            .root_source_file = b.path("src/blink.zig"),
            .imports = &.{
                .{ .name = "ir", .module = ir_dep.module("ir") },
            },
        },
    );

    mb.install_firmware(firmware, .{});
    mb.install_firmware(firmware, .{ .format = .elf });
    mb.install_firmware(test_firmware, .{});
    mb.install_firmware(test_firmware, .{ .format = .elf });
}
