import { useEffect, useState } from "react";
import type {
	LatLng,
	Position,
	WebSocketMessage,
	CurrentLocationMessage,
	RideStartedMessage,
	RidePositionAddedMessage,
	RideEndedMessage,
	RideSummary,
} from "../types";
import { useWebSocket } from "./websocket";

interface RideState {
	currentRide: RideSummary | null;
	ridePositions: Position[];
}

export function useLatLngList(url: string) {
	const { lastMessage, readyState, sendMessage } = useWebSocket(url);
	const [latLngList, setLatLngList] = useState<LatLng[]>([]);
	const [currentLocation, setCurrentLocation] = useState<Position | null>(null);
	const [rideState, setRideState] = useState<RideState>({
		currentRide: null,
		ridePositions: [],
	});

	useEffect(() => {
		if (lastMessage) {
			try {
				const message = JSON.parse(lastMessage) as WebSocketMessage;

				switch (message.type) {
					case "current_location": {
						const locationMsg = message as CurrentLocationMessage;
						const position: Position = {
							latitude: locationMsg.payload.latitude,
							longitude: locationMsg.payload.longitude,
							timestamp: locationMsg.payload.timestamp,
							speed_knots: locationMsg.payload.speed_knots,
						};

						setCurrentLocation(position);

						// Add to latLngList for live tracking
						setLatLngList((prevList) => [
							...prevList,
							{ lat: position.latitude, lng: position.longitude },
						]);
						break;
					}

					case "ride_started": {
						const rideMsg = message as RideStartedMessage;
						const newRide: RideSummary = {
							id: rideMsg.payload.ride_id,
							name: rideMsg.payload.name,
							start_time: rideMsg.payload.start_time,
						};

						setRideState({
							currentRide: newRide,
							ridePositions: [rideMsg.payload.initial_position],
						});

						// Clear previous live tracking data for new ride
						setLatLngList([
							{
								lat: rideMsg.payload.initial_position.latitude,
								lng: rideMsg.payload.initial_position.longitude,
							},
						]);
						break;
					}

					case "ride_position_added": {
						const positionMsg = message as RidePositionAddedMessage;
						setRideState((prev) => ({
							...prev,
							ridePositions: [
								...prev.ridePositions,
								positionMsg.payload.position,
							],
						}));
						break;
					}

					case "ride_ended": {
						const endMsg = message as RideEndedMessage;
						setRideState((prev) => ({
							...prev,
							currentRide: prev.currentRide
								? {
										...prev.currentRide,
										end_time: endMsg.payload.end_time,
									}
								: null,
						}));
						break;
					}

					default:
						// Fallback for old message format (backward compatibility)
						const data = message as any;
						if (
							typeof data.latitude === "number" &&
							typeof data.longitude === "number"
						) {
							setLatLngList((prevList) => [
								...prevList,
								{ lat: data.latitude, lng: data.longitude },
							]);
						} else {
							console.warn(
								"[useLatLngList] Received unknown message type:",
								message,
							);
						}
				}
			} catch (error) {
				console.error("[useLatLngList] Error parsing message:", error);
			}
		}
	}, [lastMessage]);

	return {
		latLngList,
		readyState,
		sendMessage,
		currentLocation,
		rideState,
	};
}
