const std = @import("std");
const microzig = @import("microzig");
const ir = @import("ir");

const rp2xxx = microzig.hal;
const rptime = rp2xxx.time;
const time = microzig.drivers.time;
const gpio = rp2xxx.gpio;

const led = gpio.num(25);
const button = gpio.num(12);
const ir_input = gpio.num(15);
const pwr = gpio.num(20);
const on_led = gpio.num(17);
const off_led = gpio.num(16);

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
            ir_input => {
                t2 = rptime.get_time_since_boot();
                if (parser.put(t2.diff(t1).to_us())) {
                    checkIR();
                }
                t1 = t2;
            },
            else => {},
        }
    }
}

fn togglePower() void {
    const st = pwr.read();
    pwr.put(1 - st);
    on_led.put(1 - st);
    off_led.put(st);
    blink(2);
}

fn checkIR() void {
    if (parser.value()) |msg| {
        if (msg.address == 0x35 and msg.command == 0x40) { // minidsp flex on/off button
            togglePower();
        }
    } else |_| {
        blink(5);
    }
}

fn blink(n: u4) void {
    for (0..n * 2) |_| {
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

    ir_input.set_function(.sio);
    ir_input.set_direction(.in);
    ir_input.set_pull(.up);
    ir_input.set_irq_enabled(gpio.IrqEvents{ .fall = 1, .rise = 1 }, true);

    microzig.interrupt.enable(.IO_IRQ_BANK0);

    led.set_function(.sio);
    led.set_direction(.out);
    on_led.set_function(.sio);
    on_led.set_direction(.out);
    off_led.set_function(.sio);
    off_led.set_direction(.out);
    pwr.set_function(.sio);
    pwr.set_direction(.out);

    tx_pin.set_function(.uart);

    uart.apply(.{
        .baud_rate = baud_rate,
        .clock_config = rp2xxx.clock_config,
    });

    rp2xxx.uart.init_logger(uart);
}
