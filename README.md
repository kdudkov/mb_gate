# Modbus gate

## 4-relay plate

### registers
Address | Description
---
0x4000 | device address (read/write)
0x04 | firmware month
0x08 | firmware year
0x10 | firmware hour, minute
0x20 | pcb version

### coils
Address | Description
---
0x01 | relay1 (0x100 = on?)
0x02 | relay2 (0x100 = on?)
0x03 | relay3 (0x100 = on?)
0x04 | relay4 (0x100 = on?)
0xff | all (0xffff = on?)

## din-rail plate

### registers
Address | Description
---
0x1 | device address

### coils
Address | Description
---
0x10 | relay1
0x11 | relay2
0x12 | relay3
0x13 | relay4

