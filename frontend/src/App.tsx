import "./App.css";
import { useWS } from "./hooks/useWS";
function App() {
	const { lastMessage, readyState } = useWS();
	return (
		<>
			<h1>WebSocket Example</h1>
			<p>Last message: {lastMessage || "No messages yet"}</p>
			<p>Ready state: {readyState}</p>
		</>
	);
}

export default App;
