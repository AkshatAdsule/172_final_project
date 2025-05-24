import type React from "react";
import type { ReactNode } from "react";
import { WSContext } from "../contexts/ws-context";
import { useWebSocket } from "../hooks/websocket";

interface WSProviderProps {
	url: string;
	children: ReactNode;
}

export const WSProvider: React.FC<WSProviderProps> = ({ url, children }) => {
	const ws = useWebSocket(url);
	return <WSContext.Provider value={ws}>{children}</WSContext.Provider>;
};
export { WSContext };
