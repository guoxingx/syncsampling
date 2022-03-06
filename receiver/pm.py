# coding: utf-8

import os
import time
import datetime


# 积分时间，每次采样会先跳过这段时间，再开始采集数据
INTERGRATION_TIME = 1.5

# 采集时间，每次采集会取这个长度的时间段
SAMPLING_TIME = 0.5

# 允许误差，超过这个值视为不同数据
ALLOW_ERROR = 0.01

# 与前 n 个数的均值对比，不应超过允许误差
AVG_ERROR_COUNT = 5

# 输出结果的文件名
OUTPUT_FILENAME = "output.txt"


FileName = None # filename of the output .txt file of power meter
PMActiveTs = None # activate timestamp of power meter

Prev = None


class Sampling(object):
    def __init__(self):
        self.begin = 0
        self.end = 0

        self.valid_begin = 0
        self.valid_end = 0

        self.ts_begin = 0
        self.ts_end = 0

        self.values = []
        self.valid_values = []
        self.avg = 0

    def is_stable(self, f):
        if len(self.values) == 0:
            return False

        count = AVG_ERROR_COUNT
        if len(self.values) < AVG_ERROR_COUNT:
            count = len(self.values)

        avg = sum([v for v in self.values]) / len(self.values)
        last = self.values[-1]

        err1 = abs(last - f) / last
        err2 = abs(avg - f) / avg

        if err1 > ALLOW_ERROR or err2 > ALLOW_ERROR:
            return False
        return True

        # if len(self.values) == 0:
        #     return False

        # err = abs(self.values[-1] - f)
        # if err / self.values[-1] > ALLOW_ERROR:
        #     return False
        # return True


def sampling():
    sampling_from_powermeter()


def require_filename():
    name = input("type in 'filename' of power meter: ")

    f = open(name, 'r')
    f.close()

    global FileName
    FileName = name
    return True


def empty_samping():
    """
    ;First Pulse Arrived : 06/03/2022 at 11:24:59
    """
    global Prev
    cs = Sampling()
    cs.begin = 0

    now = time.time()
    # now = 1646537100.0
    cs.ts_begin = now + INTERGRATION_TIME
    cs.ts_end = now + INTERGRATION_TIME + SAMPLING_TIME
    print("collect data between {} ~ {}".format(cs.ts_begin, cs.ts_end))

    time.sleep(INTERGRATION_TIME + SAMPLING_TIME)

    filename = "sample.txt"
    f = open(filename, 'r')

    begin = False
    for i, line in enumerate(f.readlines()):
        if "First Pulse Arrived" in line:
            # record the activate of pm
            parts = line.split("Arrived :")
            tStr = parts[1]
            parts = tStr.split("at")

            dateStr = parts[0]
            timeStr = parts[1]
            dateStr = dateStr.strip()
            timeStr = timeStr.strip()
            t = datetime.datetime.strptime("{} {}".format(dateStr, timeStr), "%d/%m/%Y %H:%M:%S")

            ts = time.mktime(t.timetuple())
            print("find activate time is {}, ts: {}".format(t, ts))

            global PMActiveTs
            PMActiveTs = ts
            continue

        if "Timestamp" in line:
            print("find the beginning of data")
            begin = True
            continue

        if not begin:
            continue

        s = line.strip()
        parts = s.split(" ")

        d, vs = float(parts[0]), parts[-1]
        # print("ts of this line: {}".format(d + PMActiveTs))
        if d + PMActiveTs < cs.ts_begin:
            # only record after INTERGRATION_TIME
            continue

        if d + PMActiveTs > cs.ts_end:
            break

        # record the first and the last line
        if cs.begin == 0:
            cs.begin = i
        cs.end = i

        value = parse_value(vs)
        cs.values.append(value)

        # only recorded when value is stabled
        if cs.is_stable(value):
            cs.valid_values.append(value)
            if cs.valid_begin == 0:
                cs.valid_begin = i
            cs.valid_end = i

    if len(cs.valid_values) == 0:
        print(cs.values)
        f.close()
        raise ValueError("falied to load valid value")

    f.close()

    cs.avg = sum([v for v in cs.valid_values]) / len(cs.valid_values)
    print("count: {}, valid_values: {}".format(len(cs.valid_values), cs.valid_values))
    print("valid line {} to {}, recorded line from {} to {}".format(cs.valid_begin, cs.valid_end, cs.begin, cs.end))
    print("avg: ", cs.avg)
    print("")

    f = open(OUTPUT_FILENAME, 'a+')
    f.write("\n")
    f.write("[采集结果] 开始时间：{}\n".format(datetime.datetime.fromtimestamp(now)))
    f.write("{}\n".format(cs.avg))
    f.close()

    global Prev
    Prev = cs


def sampling_from_powermeter():
    if Prev == None:
        print("first sampling")
        return empty_samping()

    # record timestamp in the begining
    cs = Sampling()
    cs.ts_begin = time.time()

    now = time.time()
    # now = 1646537103.0
    cs.ts_begin = now + INTERGRATION_TIME
    cs.ts_end = now + INTERGRATION_TIME + SAMPLING_TIME
    print("collect data between {} ~ {}".format(cs.ts_begin, cs.ts_end))

    time.sleep(INTERGRATION_TIME + SAMPLING_TIME)

    filename = "sample.txt"
    f = open(filename, 'r')

    print("prev samping ended in line {}".format(Prev.end))
    for i, line in enumerate(f.readlines()):
        # print(i, line)

        if i < Prev.end:
            continue

        s = line.strip()
        parts = s.split(" ")

        d, vs = float(parts[0]), parts[-1]
        # print("ts of this line: {}".format(d + PMActiveTs))
        if d + PMActiveTs < cs.ts_begin:
            # only record after INTERGRATION_TIME
            continue

        if d + PMActiveTs > cs.ts_end:
            break

        # record the first and the last line
        if cs.begin == 0:
            cs.begin = i
        cs.end = i

        value = parse_value(vs)
        cs.values.append(value)

        # only recorded when value is stabled
        if cs.is_stable(value):
            cs.valid_values.append(value)
            if cs.valid_begin == 0:
                cs.valid_begin = i
            cs.valid_end = i

    if len(cs.valid_values) == 0:
        print(cs.values)
        f.close()
        raise ValueError("falied to load valid value")

    f.close()

    cs.avg = sum([v for v in cs.valid_values]) / len(cs.valid_values)
    print("count: {}, valid_values: {}".format(len(cs.valid_values), cs.valid_values))
    print("valid line {} to {}, recorded line from {} to {}".format(cs.valid_begin, cs.valid_end, cs.begin, cs.end))
    print("avg: ", cs.avg)
    print("")

    f = open(OUTPUT_FILENAME, 'a+')
    f.write("{}\n".format(cs.avg))
    f.close()


def parse_value(s):
    return float(s)


if __name__ == "__main__":
    sampling_from_powermeter()
    sampling_from_powermeter()
