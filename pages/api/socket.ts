// Import required modules
import { NextApiRequest} from "next";
import { Server as NetServer } from "http";
import { Server as ServerIO } from "socket.io";
import { NextApiResponseServerIO } from "@/types/next";

// Initialize user count
let userCount = 0;

// Function to handle WebSocket connection
const handleSocketConnection = (io: ServerIO) => {
  io.on("connection", (socket) => {
    // Increment user count when a new user connects
    userCount++;
    io.emit("userConnect", userCount);

    // Decrement user count and notify clients when a user disconnects
    socket.on("disconnect", () => {
      userCount--;
      io.emit("userDisconnect", userCount);
    });
  });
};

export const config = {
  api: {
    bodyParser: false,
  },
};

const io = async (req: NextApiRequest, res: NextApiResponseServerIO) => {
  if (!res.socket.server.io) {
    const path = "/api/socket";
    console.log("Socket server initiated");

    // Due to NextJs 13.2.5+ not working with web sockets
    // we must default to http to send and receive messages
    // Github NextJs issue: #49334
    const httpServer: NetServer = res.socket.server as any;
    const io = new ServerIO(httpServer, {
      path: path,
      addTrailingSlash: false,
      transports: ["websocket", "polling"],
      pingInterval: 1000, // Send a ping every 1 second
      pingTimeout: 5000, // Wait for 5 seconds for a response
      perMessageDeflate: true,
    });

    // Call the function to handle WebSocket connection
    handleSocketConnection(io);

    res.socket.server.io = io;
  }
  res.end();
};

export default io;
