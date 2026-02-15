import os
BT = chr(96)
DQ = chr(34)

def jt(j):
    return BT + chr(106) + chr(115) + chr(111) + chr(110) + chr(58) + DQ + j + DQ + BT

def bjt(b, j):
    return BT + chr(98)+chr(115)+chr(111)+chr(110)+chr(58)+DQ + b + DQ + chr(32) + chr(106)+chr(115)+chr(111)+chr(110)+chr(58)+DQ + j + DQ + BT

def ft(f, j):
    return BT + chr(102)+chr(111)+chr(114)+chr(109)+chr(58)+DQ + f + DQ + chr(32) + chr(106)+chr(115)+chr(111)+chr(110)+chr(58)+DQ + j + DQ + BT

print(jt(chr(116)+chr(101)+chr(115)+chr(116)))