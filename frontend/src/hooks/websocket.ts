import { useCallback, useEffect, useRef, useState } from "react";

export type ReadyStateString = "CONNECTING" | "OPEN" | "CLOSING" | "CLOSED";

const ReadyState: Record<number, ReadyStateString> = {
	[WebSocket.CONNECTING]: "CONNECTING",
	[WebSocket.OPEN]: "OPEN",
	[WebSocket.CLOSING]: "CLOSING",
	[WebSocket.CLOSED]: "CLOSED",
};

export interface UseWebSocketResult {
	sendMessage: (
		data: string | ArrayBufferLike | Blob | ArrayBufferView,
	) => void;
	lastMessage: string | null;
	readyState: ReadyStateString;
}

export function useWebSocket(url: string): UseWebSocketResult {
	const socketRef = useRef<WebSocket | null>(null);
	const [lastMessage, setLastMessage] = useState<string | null>(null);
	const [readyState, setReadyState] = useState<ReadyStateString>("CLOSED");

	const sendMessage = useCallback(
		(data: string | ArrayBufferLike | Blob | ArrayBufferView) => {
			const socket = socketRef.current;
			if (socket && socket.readyState === WebSocket.OPEN) {
				socket.send(data);
			} else {
				console.warn(
					"[useWebSocket] cannot send, socket not open:",
					socket ? ReadyState[socket.readyState] : "NO_SOCKET",
				);
			}
		},
		[],
	);

	useEffect(() => {
		const socket = new WebSocket(url);
		socketRef.current = socket;
		setReadyState(ReadyState[socket.readyState]);

		const handleOpen = (): void => setReadyState(ReadyState[socket.readyState]);
		const handleMessage = (e: MessageEvent): void =>
			setLastMessage(e.data as string);
		const handleClose = (): void =>
			setReadyState(ReadyState[socket.readyState]);
		const handleError = (e: Event): void =>
			console.error("[WebSocket error]", e);

		socket.addEventListener("open", handleOpen);
		socket.addEventListener("message", handleMessage);
		socket.addEventListener("close", handleClose);
		socket.addEventListener("error", handleError);

		return () => {
			socket.removeEventListener("open", handleOpen);
			socket.removeEventListener("message", handleMessage);
			socket.removeEventListener("close", handleClose);
			socket.removeEventListener("error", handleError);
			socket.close();
		};
	}, [url]);

	return { sendMessage, lastMessage, readyState };
}
