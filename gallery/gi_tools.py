# coding: utf-8

import os
import time
import json
import socket
import functools
import socketserver
from datetime import datetime
from http.server import HTTPServer, SimpleHTTPRequestHandler, BaseHTTPRequestHandler


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

HOST = ("0.0.0.0", 3000)
INTERVAL = 2.5

Index = 0
Total = 0
Images = None
LastTS = 0


class MainHandler(SimpleHTTPRequestHandler):
    def do_GET(self):
        print(self.path)
        # if self.path == '/':
        #     self.path = 'index.html'

        if "/api/action" in self.path:
            return self.handle_API_action()

        elif self.path == "/api/image":
            return self.handle_API_image()

        elif self.path == "/api/images":
            return self.handle_API_images()

        elif self.path == "/api/images/contrast":
            return self.handle_API_contrast()

        return SimpleHTTPRequestHandler.do_GET(self)

    def handle_API_contrast(self):
        res = [
            "http://localhost:3000/contrast/p_1.png",
            "http://localhost:3000/contrast/p_2.png",
            "http://localhost:3000/contrast/p_3.png",
            "http://localhost:3000/contrast/p_4.png",
        ]
        self.set_API_header()
        return self.response(res)

    def handle_API_action(self):
        print("receive image ready")

        global Index
        if Index < Total:
            # record ts
            record_signal(time.time(), Index)

            self.set_API_header()
            return self.response(None)

    def handle_API_image(self):
        images = load_images()
        if len(images) <= int(Index):
            return self.response("")
        else:
            print("showing image with index {}".format(Index))

            # update index after 3s
            update_index()

            url = "http://localhost:3000/images/{}".format(images[Index])
            self.set_API_header()
            return self.response(url)

    def handle_API_images(self):
        images = load_images()
        print("images: {}".format(images))
        self.set_API_header()
        return self.response(images)

    def set_API_header(self):
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.end_headers()

    def response(self, data):
        res = {
            "code": 0,
            "data": data,
            "error": ""
        }
        self.wfile.write(json.dumps(res).encode())


def load_images():
    global Images
    if Images != None:
        return Images

    Images = []
    dir_path = "images"
    for file in os.listdir(dir_path):
        if file.endswith(".jpg") or file.endswith(".png"):
            Images.append(file)

    global Total
    Total = len(Images)

    Images.sort(key=functools.cmp_to_key(cmp_image_name))
    return Images


def cmp_image_name(a, b):
    return int(a[2:-4]) - int(b[2:-4])


def record_signal(ts, index):
    f = open("signals.txt".format(), 'a+')

    if index == 0:
        f.write("\n我是分割线\n")

    f.write("{}\n".format(ts))
    f.close()
    print("signal time {} write into file".format(ts2time(ts)))


def ts2time(ts):
    return datetime.fromtimestamp(ts).strftime("%H:%M:%S.%f")


def update_index():
    global LastTS, Index
    if LastTS == 0:
        LastTS = time.time()
    if LastTS + INTERVAL < time.time():
        Index += 1
        LastTS = time.time()


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
        name = input("输入文件名: ")
    else:
        name = input("输入{}: ".format(note))

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

    log("line before {} will be skiped".format(Prev.end))
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


def mode_required():
    print("鬼像数据采集v1.0")
    print("选择要执行的功能:")
    print("    1 - 数据提取(默认)")
    print("    2 - 散斑图投影")
    mode = input("输入 1 或 2: ")

    if mode == "":
        mode = 1
    else:
        mode = int(mode)
    return mode


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


if __name__ == "__main__":
    mode = mode_required()

    if mode == 1:
        signalfile= filename_required("signals.txt", "散斑信号文件(默认signals.txt)")
        pm_out = filename_required("sample.txt", "功率计文件(默认sample.txt)")

        f = open(OUTPUT_FILENAME, 'a+')
        tss = ts_from_recorded_signals(signalfile)
        for i, ts in enumerate(tss):
            res = sampling(ts, pm_out, i)
            f.write("{}\n".format(res))
        f.close()

    elif mode == 2:
        log("---------------------------------")
        log("启动散斑图投影系统")
        log("请打开浏览器localhost:{}".format(HOST[1]))
        log("开始投影前请确保starlab已开始记录")
        log("---------------------------------")
        log("投影完成后，用starlab记录文件和signals.txt来提取数据")
        log("---------------------------------")
        handler_object = MainHandler
        server = socketserver.TCPServer(HOST, handler_object)
        server.serve_forever()

    else:
        print("无效指令")
        exit()
