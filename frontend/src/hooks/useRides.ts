import { useEffect, useState } from "react";
import type {
	RideSummary,
	RideDetail,
	WebSocketMessage,
	RideStartedMessage,
	RideEndedMessage,
	RidePositionUpdateMessage,
} from "../types";
import { ApiService } from "../services/api";
import { useWS } from "./useWS";

export function useRides() {
	const [rides, setRides] = useState<RideSummary[]>([]);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState<string | null>(null);
	const { lastMessage } = useWS();

	const fetchRides = async (page = 1, limit = 50, date?: string) => {
		try {
			setLoading(true);
			setError(null);
			const ridesData = await ApiService.getRides(page, limit, date);
			setRides(ridesData);
		} catch (err) {
			setError(err instanceof Error ? err.message : "Failed to fetch rides");
			console.error("Error fetching rides:", err);
		} finally {
			setLoading(false);
		}
	};

	useEffect(() => {
		fetchRides();
	}, []);

	useEffect(() => {
		if (lastMessage) {
			try {
				const message = JSON.parse(lastMessage) as WebSocketMessage;
				switch (message.type) {
					case "RIDE_STARTED": {
						const rideMsg = message as RideStartedMessage;
						const newRide: RideSummary = {
							id: rideMsg.payload.ride_id,
							name: rideMsg.payload.ride_name,
							start_time: rideMsg.payload.timestamp,
						};
						setRides((prevRides) => [newRide, ...prevRides]);
						break;
					}
					case "RIDE_ENDED": {
						const endMsg = message as RideEndedMessage;
						setRides((prevRides) =>
							prevRides.map((ride) =>
								ride.id === endMsg.payload.ride_id
									? { ...ride, end_time: endMsg.payload.timestamp }
									: ride,
							),
						);
						break;
					}
				}
			} catch (err) {
				console.error("Failed to parse WebSocket message in useRides", err);
			}
		}
	}, [lastMessage]);

	return {
		rides,
		loading,
		error,
		refetch: fetchRides,
	};
}

export function useRideDetail(rideId: number | null) {
	const [rideDetail, setRideDetail] = useState<RideDetail | null>(null);
	const [loading, setLoading] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const { lastMessage } = useWS();

	useEffect(() => {
		if (rideId === null) {
			setRideDetail(null);
			return;
		}

		const fetchRideDetail = async () => {
			try {
				setLoading(true);
				setError(null);
				const detail = await ApiService.getRideDetail(rideId);
				setRideDetail(detail);
			} catch (err) {
				setError(
					err instanceof Error ? err.message : "Failed to fetch ride detail",
				);
				console.error("Error fetching ride detail:", err);
			} finally {
				setLoading(false);
			}
		};

		fetchRideDetail();
	}, [rideId]);

	useEffect(() => {
		if (lastMessage && rideId !== null) {
			try {
				const message = JSON.parse(lastMessage) as WebSocketMessage;
				switch (message.type) {
					case "RIDE_POSITION_UPDATE": {
						const posMsg = message as RidePositionUpdateMessage;
						if (posMsg.payload.ride_id === rideId) {
							setRideDetail((prev) => {
								if (!prev) return null;
								return {
									...prev,
									positions: [...prev.positions, posMsg.payload.position],
								};
							});
						}
						break;
					}
					case "RIDE_ENDED": {
						const endMsg = message as RideEndedMessage;
						if (endMsg.payload.ride_id === rideId) {
							setRideDetail((prev) => {
								if (!prev) return null;
								return {
									...prev,
									end_time: endMsg.payload.timestamp,
								};
							});
						}
						break;
					}
				}
			} catch (err) {
				console.error(
					"Failed to parse WebSocket message in useRideDetail",
					err,
				);
			}
		}
	}, [lastMessage, rideId]);

	return {
		rideDetail,
		loading,
		error,
	};
}
