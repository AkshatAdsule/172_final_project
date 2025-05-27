import { useEffect, useState } from "react";
import type { LatLng } from "../types";
import { useWebSocket } from "./websocket";

export function useLatLngList(url: string) {
	const { lastMessage, readyState, sendMessage } = useWebSocket(url);
	const [latLngList, setLatLngList] = useState<LatLng[]>([]);

	useEffect(() => {
		if (lastMessage) {
			try {
				const data = JSON.parse(lastMessage) as {
					latitude: number;
					longitude: number;
				};
				// Basic validation to ensure lat and lng are numbers
				if (
					typeof data.latitude === "number" &&
					typeof data.longitude === "number"
				) {
					setLatLngList((prevList) => [
						...prevList,
						{ lat: data.latitude, lng: data.longitude },
					]);
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
