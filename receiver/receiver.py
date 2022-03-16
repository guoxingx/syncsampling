# coding: utf-8

import os
import time
import json
import socket
from datetime import datetime


HostAddr = None
HostPort = 26001

# 积分时间，每次采样会先跳过这段时间，再开始采集数据
INTERGRATION_TIME = 1.5

# 采集时间，每次采集会取这个长度的时间段
SAMPLING_TIME = 0.5

# 允许误差，超过这个值视为不同数据
ALLOW_ERROR = 0.1

# 输出结果的文件名
OUTPUT_FILENAME = "output.txt"
SIGNAL_FILE = "signals.txt"

SIMU_TIME = 1646622842
SIMU_D = 0

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

        last = self.values[-1]

        err1 = abs(last - f) / last

        if err1 > ALLOW_ERROR:
            return False
        return True

        # if len(self.values) == 0:
        #     return False

        # err = abs(self.values[-1] - f)
        # if err / self.values[-1] > ALLOW_ERROR:
        #     return False
        # return True


def ts2time(ts):
    return datetime.fromtimestamp(ts).strftime("%H:%M:%S.%f")


def sampling(ts, pm_output_file, index):
    # try:
    #     return sampling_from_powermeter(ts, pm_output_file)
    #     log("image at index {} sampling complete".format(index))
    # except Exception as e:
    #     log("sampling failed with exception: {}".format(e))
    #     return 0

    return sampling_from_powermeter(ts, pm_output_file)
    log("image at index {} sampling complete".format(index))
    return 0


def filename_required(default=None, note=None):
    if note == None:
        name = input("type in 'filename': ")
    else:
        name = input("type in 'filename' of {}: ".format(note))

    if name == "" and default != None:
        name = default

    f = open(name, 'r')
    f.close()
    return name


def empty_samping(ts, pm_output_file):
    """
    ;First Pulse Arrived : 06/03/2022 at 11:24:59
    """
    cs = Sampling()
    cs.begin = 0

    cs.ts_begin = ts + INTERGRATION_TIME
    cs.ts_end = ts + INTERGRATION_TIME + SAMPLING_TIME
    log("collect data between {} ~ {}".format(ts2time(cs.ts_begin), ts2time(cs.ts_end)))

    # time.sleep(INTERGRATION_TIME + SAMPLING_TIME)

    f = open(pm_output_file, 'r')

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
            t = datetime.strptime("{} {}".format(dateStr, timeStr), "%d/%m/%Y %H:%M:%S")

            ts = time.mktime(t.timetuple())
            log("find activate time is {}, time: {}, line: {}".format(t, ts2time(ts), i))

            global PMActiveTs
            PMActiveTs = ts
            continue

        if "Timestamp" in line:
            log("find the beginning of data, line: {}".format(i))
            begin = True
            continue

        if not begin:
            continue

        s = line.strip()
        parts = s.split(" ")

        d, value = parts[0], parts[-1]
        try:
            d = float(d)
            value = float(value)
        except Exception as e:
            log("failed to convert interval: {}".format(e))
            log("line {}: {} into {}".format(i, line, parts))

        if d + PMActiveTs < cs.ts_begin:
            # only record after INTERGRATION_TIME
            continue

        if d + PMActiveTs > cs.ts_end:
            log("out of range of time, loop will be finished")
            break

        log("[valid] value: {}, line: {}, time: {}".format(value, i, ts2time(d + PMActiveTs)))
        # record the first and the last line
        if cs.begin == 0:
            cs.begin = i
        if i > cs.end:
            cs.end = i

        cs.values.append(value)

        # only recorded when value is stabled
        if cs.is_stable(value):
            cs.valid_values.append(value)
            if cs.valid_begin == 0:
                cs.valid_begin = i
            cs.valid_end = i

    log("finish loop, ended in line {}".format(i))
    if len(cs.valid_values) == 0:
        log("no valid values found!")
        log("all values: {}".format(cs.values))
        f.close()

    f.close()

    if len(cs.valid_values) == 0:
        log("falied to load valid value, 0 will be write into output file")
        cs.avg = 0
    else:
        cs.avg = sum([v for v in cs.valid_values]) / len(cs.valid_values)
        log("valid line {} to {}, recorded line from {} to {}".format(cs.valid_begin, cs.valid_end, cs.begin, cs.end))
        log("avg: {}, count: {}, valid_values: {}".format(cs.avg, len(cs.valid_values), cs.valid_values))
        print("")

    f = open(OUTPUT_FILENAME, 'a+')
    f.write("\n")
    f.write("[采集结果] 开始时间：{}\n".format(datetime.fromtimestamp(ts)))
    f.close()

    global Prev
    Prev = cs

    return cs.avg


def sampling_from_powermeter(ts, pm_output_file):
    global Prev
    if Prev == None:
        log("first sampling")
        return empty_samping(ts, pm_output_file)

    # record timestamp in the begining
    cs = Sampling()
    cs.ts_begin = time.time()
    cs.end = Prev.end

    cs.ts_begin = ts + INTERGRATION_TIME
    cs.ts_end = ts + INTERGRATION_TIME + SAMPLING_TIME
    log("collect data between time {} ~ {}".format(ts2time(cs.ts_begin), ts2time(cs.ts_end)))

    # time.sleep(INTERGRATION_TIME + SAMPLING_TIME)

    f = open(pm_output_file, 'r')

    log("line befor {} will be skiped".format(Prev.end))
    for i, line in enumerate(f.readlines()):
        if i < Prev.end:
            continue

        s = line.strip()
        parts = s.split(" ")

        d, value = parts[0], parts[-1]
        try:
            d = float(d)
            value = float(value)
        except Exception as e:
            log("failed to convert interval: {}".format(e))
            log("line {}: {} into {}".format(i, line, parts))

        if d + PMActiveTs < cs.ts_begin:
            # only record after INTERGRATION_TIME
            continue

        if d + PMActiveTs > cs.ts_end:
            log("out of range of time, loop will be finished")
            break

        # record the first and the last line
        if cs.begin == 0:
            cs.begin = i
        if i > cs.end:
            cs.end = i

        log("[valid] value: {}, line: {}, time: {}".format(value, i, ts2time(d + PMActiveTs)))
        cs.values.append(value)

        # only recorded when value is stabled
        if cs.is_stable(value):
            cs.valid_values.append(value)
            if cs.valid_begin == 0:
                cs.valid_begin = i
            cs.valid_end = i

    log("finish loop, ended in line {}".format(i))
    if len(cs.valid_values) == 0:
        log("no valid values found!")
        log("all values: {}".format(cs.values))
        f.close()

    f.close()

    if len(cs.valid_values) == 0:
        log("falied to load valid value, 0 will be write into output file")
        cs.avg = 0
    else:
        cs.avg = sum([v for v in cs.valid_values]) / len(cs.valid_values)
        log("valid line {} to {}, recorded line from {} to {}".format(cs.valid_begin, cs.valid_end, cs.begin, cs.end))
        log("avg: {}, count: {}, valid_values: {}".format(cs.avg, len(cs.valid_values), cs.valid_values))
        print("")

    Prev = None
    Prev = cs
    return cs.avg


def require_host():
    host = input("type in ip of Tx: ")
    if host == "":
        host = "localhost"
    host = host.strip()

    global HostAddr
    HostAddr = host
    return True


def mode_required():
    print("choose the function to be executed:")
    print("    1 - collect data from a signals file")
    print("    2 - run as a RX (connecting to TX required)")
    mode = input("type in 1 or 2: ")

    if mode == "":
        mode = 1
    else:
        mode = int(mode)
    return mode


def record_signal(ts):
    f = open("signals.txt".format(), 'a+')
    f.write("{}\n".format(ts))
    f.close()
    log("signal time {} write into file".format(ts2time(ts)))


def ts_from_recorded_signals(signalfile):
    tss = []
    f = open(signalfile, 'r')
    for i, line in enumerate(f.readlines()):
        ts = float(line)
        if ts > 1000000000000:
            ts = ts / 1000
        tss.append(ts)
    f.close()
    return tss


def log(s):
    now = datetime.now()
    # print("{}: {}".format(now.strftime("%Y-%m-%d %H:%M:%S.%f"), s))
    print("{}: {}".format(now, s))


def start_client(pm_output_file):
    log("try to connect tcp host in {}:{}".format(HostAddr, HostPort))
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    host = (HostAddr, HostPort)
    # host = ("localhost", 26001)
    client.connect(host)
    log("tcp connect success!")

    while True:
        """
        {"i":0,"a":1,"t": 16467206834791}\n
        t: ts in millisecond
        """
        data = client.recv(255)
        log("receive raw from Tx: {}".format(data))

        message = json.loads(data.decode())
        # print("receive from Tx: {}".format(message))

        index = message["i"]
        log("image at index {} is ready for sampling".format(index))

        ts = time.time()
        if message["a"] == 4:
            ts = message["t"] / 1000.0
            log("use signal time: {}".format(datetime.fromtimestamp(ts)))

        # sampling
        res = sampling(ts, pm_output_file, index)

        # plan B
        record_signal(ts)

        f = open(OUTPUT_FILENAME, 'a+')
        f.write("{}\n".format(res))
        f.close()

        log("\n")
        # log("response to Tx\n")

        # sendMessage = {"i": index,"a": 2}
        # sendData = json.dumps(sendMessage).encode()
        # end = b'\n'
        # sendData = sendData + end
        # client.send(sendData)


if __name__ == "__main__":
    mode = mode_required()

    if mode == 1:
        signalfile= filename_required("signals.txt", "signal file")
        pm_out = filename_required("sample.txt", "power meter")

        f = open(OUTPUT_FILENAME, 'a+')
        tss = ts_from_recorded_signals(signalfile)
        for i, ts in enumerate(tss):
            res = sampling(ts, "sample.txt", i)
            f.write("{}\n".format(res))
        f.close()

    elif mode == 2:
        if not require_host():
            print("unrecognized host, abort")
            exit()

        pm_out = filename_required("sample.txt", "power meter")
        start_client(pm_out)

    else:
        print("unrecognized mode, abort")
        exit()
