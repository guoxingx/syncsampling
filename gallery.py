# coding: utf-8

import os
import json
import time
import functools
import socketserver
from datetime import datetime
from http.server import HTTPServer, SimpleHTTPRequestHandler, BaseHTTPRequestHandler

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

        return SimpleHTTPRequestHandler.do_GET(self)

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
    dir_path = "/Users/wuyi/developer/projects/syncsampling/statics/images"
    for file in os.listdir(dir_path):
        if file.endswith(".jpg"):
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


if __name__ == "__main__":
    handler_object = MainHandler
    server = socketserver.TCPServer(HOST, handler_object)
    server.serve_forever()

