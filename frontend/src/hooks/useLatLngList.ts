import { useEffect, useState } from "react";
import type { LatLng } from "../types";
import { useWebSocket } from "./websocket";

export function useLatLngList(url: string) {
	const { lastMessage, readyState, sendMessage } = useWebSocket(url);
	const [latLngList, setLatLngList] = useState<LatLng[]>([]);

	useEffect(() => {
		if (lastMessage) {
			try {
				const data = JSON.parse(lastMessage) as LatLng;
				// Basic validation to ensure lat and lng are numbers
				if (typeof data.lat === "number" && typeof data.lng === "number") {
					setLatLngList((prevList) => [...prevList, data]);
				} else {
					console.warn("[useLatLngList] Received invalid data format:", data);
				}
			} catch (error) {
				console.error("[useLatLngList] Error parsing message:", error);
			}
		}
	}, [lastMessage]);

	return { latLngList, readyState, sendMessage };
}
