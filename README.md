## How to start

Clone the repository:

```bash
git clone git@github.com:therealyo/justdone.git
```

Run locally:

## Documentation

Link to endpoint documentation

http://localhost:8080/swagger/index.html#/

## Order Processor Logic

The **OrderProcessor** struct is responsible for handling incoming order events, ensuring the correct sequence of events, and managing the order lifecycle. Here's a explanation of its core logic:

### Event Handling and Deduplication:

- Each event is checked to see if it has already been processed or is currently being processed to avoid duplicates.
- If an event has been processed before, it is ignored to prevent redundant operations.

### Order Creation:

- If the first event received is not **cool_order_created**, the event is rejected with an error, triggering JustPay!'s retry mechanism to ensure the correct initial event is received.
- If the first event is **cool_order_created**, a new order is created in the database.

### Event Processing:

- Events are processed in a thread-safe manner using a mutex to prevent race conditions.
- Events are appended to the order’s history and sorted by the **created_at** timestamp to maintain the correct order.
- If the event sequence is valid, the order is updated, and the event is saved to the database. If the sequence is invalid, no update is made, and the event is not propagated.

### Error Handling:

- If an error occurs after the event is saved but before the order update is completed, the event is deleted from the database to allow JustPay! to resend the event.

### Finalizing Orders:

- Upon receiving the final status (chinazes), the processor starts a timer to finalize the order if no further events are received within the defined timeframe.
- The final status is confirmed, and the order is marked as complete.

### Notification:

- After a successful order update, all connected clients are notified of the new event through the OrderObserver.

- Notifications ensure that clients receive events in the correct sequence, even if the events were received out of order.

## SSE Notifier

The **SSENotifier** struct is responsible for handling the streaming of events to users
