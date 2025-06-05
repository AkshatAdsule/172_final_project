import { useEffect, useState } from "react";
import type {
	Position,
	WebSocketMessage,
	CurrentLocationMessage,
	RideStartedMessage,
	RidePositionUpdateMessage,
	RideEndedMessage,
	RideSummary,
} from "../types";
import { useWS } from "./useWS";

interface RideState {
	currentRide: RideSummary | null;
	ridePositions: Position[];
}

export function useLatLngList() {
	const { lastMessage, readyState, sendMessage } = useWS();
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
							...locationMsg.payload,
						};
						setCurrentLocation(position);
						break;
					}

					case "RIDE_STARTED": {
						const rideMsg = message as RideStartedMessage;
						const newRide: RideSummary = {
							id: rideMsg.payload.ride_id,
							name: rideMsg.payload.ride_name,
							start_time: rideMsg.payload.timestamp,
						};

						setRideState({
							currentRide: newRide,
							ridePositions: [rideMsg.payload.position],
						});
						break;
					}

					case "RIDE_POSITION_UPDATE": {
						const positionMsg = message as RidePositionUpdateMessage;
						setRideState((prev) => ({
							...prev,
							ridePositions: [
								...prev.ridePositions,
								positionMsg.payload.position,
							],
						}));
						break;
					}

					case "RIDE_ENDED": {
						const endMsg = message as RideEndedMessage;
						setRideState((prev) => ({
							...prev,
							currentRide: prev.currentRide
								? {
										...prev.currentRide,
										end_time: endMsg.payload.timestamp,
									}
								: null,
						}));
						break;
					}
					default: {
						console.warn(
							"[useLatLngList] Received unknown message type:",
							message.type,
						);
					}
				}
			} catch (error) {
				console.error("[useLatLngList] Error parsing message:", error);
			}
		}
	}, [lastMessage]);

	return {
		readyState,
		sendMessage,
		currentLocation,
		rideState,
	};
}
