const std = @import("std");
const microzig = @import("microzig");
const ir = @import("ir");

const rp2xxx = microzig.hal;
const rptime = rp2xxx.time;
const time = microzig.drivers.time;
const gpio = rp2xxx.gpio;

const button = gpio.num(9);
const irPin = gpio.num(10);
const led = gpio.num(25);

const uart = rp2xxx.uart.instance.num(0);
const tx_pin = gpio.num(0);
const baud_rate = 115200;

pub const microzig_options = microzig.Options{
    .log_level = .debug,
    .logFn = rp2xxx.uart.log,
    .interrupts = .{ .IO_IRQ_BANK0 = .{ .c = callback } },
};

var t1: time.Absolute = undefined;
var t2: time.Absolute = undefined;
var parser = ir.IR{};

fn callback() linksection(".ram_text") callconv(.c) void {
    var iter = gpio.IrqEventIter{};
    while (iter.next()) |e| {
        switch (e.pin) {
            button => {
                led.toggle();
                std.log.debug("button fall: {}, rise: {}", .{ e.events.fall, e.events.rise });
            },
            irPin => {
                t2 = rptime.get_time_since_boot();
                if (parser.put(t2.diff(t1).to_us())) {
                    if (parser.value()) |msg| {
                        if (msg.address == 0x35 and msg.command == 0x40) {
                            led.toggle();
                        }
                    } else |_| {
                        blink();
                    }
                }
                t1 = t2;
            },
            else => {},
        }
    }
}

fn blink() void {
    for (0..5) |_| {
        led.toggle();
        rptime.sleep_ms(250);
    }
}

pub fn main() !void {
    init();
    t1 = rptime.get_time_since_boot();
    while (true) {
        rptime.sleep_ms(2_000);
    }
}

fn init() void {
    button.set_function(.sio);
    button.set_direction(.in);
    button.set_pull(.down);
    button.set_irq_enabled(gpio.IrqEvents{ .fall = 0, .rise = 1 }, true);

    irPin.set_function(.sio);
    irPin.set_direction(.in);
    irPin.set_pull(.up);
    irPin.set_irq_enabled(gpio.IrqEvents{ .fall = 1, .rise = 1 }, true);

    microzig.interrupt.enable(.IO_IRQ_BANK0);

    led.set_function(.sio);
    led.set_direction(.out);

    tx_pin.set_function(.uart);

    uart.apply(.{
        .baud_rate = baud_rate,
        .clock_config = rp2xxx.clock_config,
    });

    rp2xxx.uart.init_logger(uart);
}
