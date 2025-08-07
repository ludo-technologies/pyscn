# Async/await patterns

import asyncio
from typing import List, AsyncIterator, AsyncContextManager

# Simple async function
async def simple_async():
    return "async result"

# Async function with await
async def fetch_data(url: str):
    # Simulate async operation
    await asyncio.sleep(1)
    return f"Data from {url}"

# Multiple awaits
async def multiple_awaits():
    result1 = await fetch_data("url1")
    result2 = await fetch_data("url2")
    return result1, result2

# Concurrent execution with gather
async def concurrent_fetches():
    results = await asyncio.gather(
        fetch_data("url1"),
        fetch_data("url2"),
        fetch_data("url3")
    )
    return results

# Async with timeout
async def with_timeout():
    try:
        result = await asyncio.wait_for(fetch_data("url"), timeout=5.0)
        return result
    except asyncio.TimeoutError:
        return "Timeout occurred"

# Async iterator
class AsyncRange:
    def __init__(self, start, stop):
        self.start = start
        self.stop = stop
    
    def __aiter__(self):
        self.current = self.start
        return self
    
    async def __anext__(self):
        if self.current < self.stop:
            await asyncio.sleep(0.1)
            value = self.current
            self.current += 1
            return value
        raise StopAsyncIteration

# Using async iterator
async def use_async_iterator():
    async for i in AsyncRange(0, 5):
        print(i)

# Async generator
async def async_generator(n):
    for i in range(n):
        await asyncio.sleep(0.1)
        yield i

# Using async generator
async def use_async_generator():
    async for value in async_generator(5):
        print(value)

# Async context manager
class AsyncContextManager:
    async def __aenter__(self):
        print("Entering async context")
        await asyncio.sleep(0.1)
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        print("Exiting async context")
        await asyncio.sleep(0.1)

# Using async context manager
async def use_async_context():
    async with AsyncContextManager() as ctx:
        print("Inside async context")
        await asyncio.sleep(0.1)

# Async with multiple context managers
async def multiple_async_contexts():
    async with AsyncContextManager() as ctx1, AsyncContextManager() as ctx2:
        print("Inside multiple contexts")

# Async comprehensions
async def async_comprehensions():
    # List comprehension with async
    results = [await fetch_data(f"url{i}") for i in range(3)]
    
    # Async for in comprehension
    async_results = [x async for x in async_generator(5)]
    
    return results, async_results

# Task creation and management
async def create_tasks():
    # Create tasks
    task1 = asyncio.create_task(fetch_data("url1"))
    task2 = asyncio.create_task(fetch_data("url2"))
    
    # Wait for tasks
    result1 = await task1
    result2 = await task2
    
    return result1, result2

# Async queue
async def producer(queue: asyncio.Queue):
    for i in range(5):
        await asyncio.sleep(0.5)
        await queue.put(f"item_{i}")

async def consumer(queue: asyncio.Queue):
    while True:
        item = await queue.get()
        print(f"Consumed: {item}")
        queue.task_done()

async def queue_example():
    queue = asyncio.Queue()
    
    # Create producer and consumer tasks
    producer_task = asyncio.create_task(producer(queue))
    consumer_task = asyncio.create_task(consumer(queue))
    
    # Wait for producer to finish
    await producer_task
    
    # Wait for queue to be empty
    await queue.join()
    
    # Cancel consumer
    consumer_task.cancel()

# Async locks
async def protected_resource(lock: asyncio.Lock, resource_id: int):
    async with lock:
        print(f"Accessing resource {resource_id}")
        await asyncio.sleep(1)
        print(f"Done with resource {resource_id}")

async def lock_example():
    lock = asyncio.Lock()
    await asyncio.gather(
        protected_resource(lock, 1),
        protected_resource(lock, 2),
        protected_resource(lock, 3)
    )

# Exception handling in async
async def async_with_exceptions():
    try:
        result = await fetch_data("url")
        return result
    except asyncio.CancelledError:
        print("Operation cancelled")
        raise
    except Exception as e:
        print(f"Error: {e}")
        return None

# Async class methods
class AsyncClass:
    async def async_method(self):
        await asyncio.sleep(0.1)
        return "async method result"
    
    @classmethod
    async def async_class_method(cls):
        await asyncio.sleep(0.1)
        return "async class method"
    
    @staticmethod
    async def async_static_method():
        await asyncio.sleep(0.1)
        return "async static method"

# Main async function
async def main():
    # Run various async operations
    result1 = await simple_async()
    result2 = await concurrent_fetches()
    await use_async_iterator()
    await use_async_context()
    
    return "All async operations completed"

# Running async code
if __name__ == "__main__":
    asyncio.run(main())