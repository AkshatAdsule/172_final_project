import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./index.css";
import App from "./App.tsx";
import { WSProvider } from "./providers/ws.tsx";

// biome-ignore lint/style/noNonNullAssertion: Root element is guaranteed to exist
createRoot(document.getElementById("root")!).render(
	<StrictMode>
		<WSProvider url="ws://localhost:8080/ws">
			<App />
		</WSProvider>
	</StrictMode>,
);
