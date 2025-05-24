import { useContext } from "react";
import { WSContext } from "../providers/ws";
import type { UseWebSocketResult } from "./websocket";

export function useWS(): UseWebSocketResult {
	const context = useContext(WSContext);
	if (!context) {
		throw new Error("useWS must be used within a WSProvider");
	}
	return context;
}
