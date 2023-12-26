# Simple chat

A simple chat room app demonstrating a full-stack chatroom implementation.

## Using the app

1. Run the server on port 8081 (If you need to change the port, it can be changed in `main.go`. You will also need to change the value of `BASE_URL` in `client/src/hooks/useChats.ts`):

```
cd server
go run .
```

2. Run the client on port 3000 from another terminal:

```
cd client
npm ci
npm run start
```

3. A browser tab with the client should automatically open, you can also navigate to `localhost:3000` in the browser.

4. Enter any username, chats that originated from that username are shown in blue once authenticated to the chat room.

5. The correct password to authenticate to the server is currently just "`password`". Once connected, you should see the chat interface.

6. Send chats using the text box at the bottom of the page, and the `Send` button.

7. Navigate to `localhost:3000` in as many tabs as you want, and repeat steps 4-6.
