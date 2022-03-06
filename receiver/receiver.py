# coding: utf-8

import socket
import time
import json

from pm import sampling, require_filename


HostAddr = None
HostPort = None


def require_host():
    host = input("type in 'ip:port' of Tx: ")

    parts = host.split(":")
    if len(parts) != 2:
        print("invalid host: {}".format(host))
        return False

    global HostAddr, HostPort
    HostAddr, HostPort = parts[0], int(parts[1])
    return True


def start_client():
    print("try to connect tcp host in {}:{}".format(HostAddr, HostPort))
    client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    host = (HostAddr, HostPort)
    # host = ("localhost", 26001)
    client.connect(host)

    while True:
        """
        {"i":0,"a":1,"e":""}\n
        """
        data = client.recv(255)
        # print("receive raw from Tx: {}".format(data))

        message = json.loads(data.decode())
        # print("receive from Tx: {}".format(message))

        index = message["i"]
        print("image at index {} is ready for sampling".format(index))

        sampling()
        print("image at index {} sampling complete".format(index))
        print("response to Tx\n")

        sendMessage = {"i": index,"a": 2,"e": ""}
        sendData = json.dumps(sendMessage).encode()
        tag = b'201'
        end = b'\n'
        sendData = sendData + end
        # print("sending data {} to Tx".format(sendData))
        client.send(sendData)


if __name__ == "__main__":
    if not require_host():
        print("abort")
        exit()

    if not require_filename():
        print("abort")
        exit()

    start_client()
