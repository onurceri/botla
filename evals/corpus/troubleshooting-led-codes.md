# Troubleshooting and LED Codes

## LED status codes

- Slow green pulse: normal operation.
- Blinking amber: not connected to Wi-Fi. Check that the network is 2.4 GHz and that the password is correct.
- Fast blinking blue: pairing mode.
- Blinking red: low battery (below 10%).
- Solid red: sensor hardware fault. Contact support to start a replacement (RMA).

## CO2 calibration

The Beacon uses Automatic Baseline Calibration (ABC), which runs every 8 days. ABC assumes the sensor is exposed to fresh air of about 420 ppm CO2 at least once during that period. In spaces that are occupied 24/7, disable ABC in the Console and perform a manual calibration by placing the sensor outdoors for 20 minutes and starting "Calibrate CO2" from the Console.

## Factory reset

Hold the button on the back of the sensor for 12 seconds until the LED flashes white.

## Gaps in readings

If a sensor loses its connection, it buffers up to 72 hours of readings locally and uploads them automatically when the connection is restored.
