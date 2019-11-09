# coding: utf-8
import asyncio
import http
import json
import os
import threading
import time

import aiohttp
import mimesis


SLEEP_TIME = .01
LINK = 'http://localhost:8080/signup'
NUMBER_OF_REQUESTS = 1_000_000

person_queue: asyncio.Queue = asyncio.Queue()

person = mimesis.Person()
text = mimesis.Text()
successes, fails = 0, 0
lock = asyncio.Lock()


class PersonCreator:

    def __init__(self, q: asyncio.Queue):
        self.__queue = q

    @staticmethod
    def create_person() -> dict:
        password = person.password()
        return {
            'first_name': person.first_name(),
            'last_name': person.last_name(),
            'age': person.age(),
            'email': person.email(),
            'password': password,
            'password2': password,
            'city': 'Los Angeles',
        }

    def loop(self) -> None:
        for _ in range(NUMBER_OF_REQUESTS):
            person = self.create_person()
            while True:
                if self.__queue.qsize() < 100:
                    self.__queue.put_nowait(person)
                    break
                time.sleep(SLEEP_TIME)


class Controller:

    def __init__(self, source: asyncio.Queue):
        self.__source_queue = source

    async def loop(self) -> None:
        global successes, fails, lock
        while True:
            person = await self.__source_queue.get()
            self.__source_queue.task_done()
            async with aiohttp.ClientSession() as session:
                async with session.post(LINK, data=person) as response:
                    async with lock:
                        if response.status == http.HTTPStatus.OK:
                            successes += 1
                            with open('/tmp/_result.json', 'w') as wfile:
                                json.dump({
                                    'successes': successes,
                                    'fails': fails,
                                    'person': person
                                }, wfile)
                            os.rename('/tmp/_result.json', '/tmp/result.json')
                        else:
                            fails += 1


if __name__ == '__main__':
    creator = PersonCreator(person_queue)
    t = threading.Thread(target=creator.loop)
    t.start()
    loop = asyncio.get_event_loop()
    loop.run_until_complete(asyncio.gather(
        *(Controller(person_queue).loop() for _ in range(10))
    ))
