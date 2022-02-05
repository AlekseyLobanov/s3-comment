import time
import sys

import aiohttp
import asyncio

TARGET = sys.argv[1]
REQUESTS_COUNT = 30000

MARKDOWN_INPUT = """## Emphasis

**This is bold text**

__This is bold text__

*This is italic text*

_This is italic text_

~~Strikethrough~~""".strip()
MARKDOWN_INPUT = 'Hello *World*'



async def get_result(query_cor):
    resp = await query_cor
    return {"text": 4}#await resp.json()

async def aio_main():
    async with aiohttp.ClientSession(connector=aiohttp.TCPConnector(limit=150)) as session:
        responds = [get_result(session.post(TARGET, json={'text': MARKDOWN_INPUT})) for _ in range(REQUESTS_COUNT)]

        responds = await asyncio.gather(*responds)
        for res in responds:
            if not res["text"]:
                print("failed conversion")
        """
        for response in responds:
            if response.status != 200:
                print(f"failed with status {response.status}")
        """

def main():
    loop = asyncio.get_event_loop()

    begin_at = time.time()
    loop.run_until_complete(aio_main())
    end_at = time.time()

    print(f"Takes {end_at - begin_at:.2f}, {REQUESTS_COUNT / (end_at - begin_at):.2f} RPS")

if __name__ == "__main__":
    main()
