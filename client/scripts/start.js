import minimist from "minimist";
import { spawn } from "child_process";

// Parse command-line arguments
const args = minimist(process.argv.slice(2), {
  string: ["user-id", "user-name", "user-photo"],
  default: {
    port: 3000,
    "user-id": "550e8400-e29b-41d4-a716-446655440000",
    "user-name": "John",
    "user-photo": "https://randomuser.me/api/portraits/men/1.jpg",
  },
});

// Set environment variables
process.env.VITE_PORT = args.port;
process.env.VITE_USER_ID = args["user-id"];
process.env.VITE_USER_NAME = args["user-name"];
process.env.VITE_USER_PHOTO = args["user-photo"];

// Start Vite with the environment variables
const vite = spawn("vite", ["dev", "--port", args.port], {
  stdio: "inherit", // Share stdout/stderr with the parent process
});

vite.on("close", (code) => {
  console.log(`Vite process exited with code ${code}`);
});
