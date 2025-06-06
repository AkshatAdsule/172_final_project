import { createFileRoute } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { MapComponent } from "../../components/map/MapComponent";
import { useLatLng } from "../../hooks/useLatLng";
import { ApiService } from "../../services/api";
// import { formatSpeedMph } from "../../utils/speed";
import type { LatLng, Position, RideDetail } from "../../types";
import styles from "./styles/live.module.css";

export const Route = createFileRoute("/rides/live")({
	component: RouteComponent,
});

function RouteComponent() {
	const { currentLocation, rideState } = useLatLng();
	const [fallbackRide, setFallbackRide] = useState<RideDetail | null>(null);
	const [loadingFallback, setLoadingFallback] = useState(false);

	// Fetch latest ride from server when needed for fallback
	useEffect(() => {
		const fetchLatestRideForFallback = async () => {
			// Only fetch if we don't have current location and don't have current ride from websocket
			if (
				!currentLocation &&
				!rideState.currentRide &&
				!fallbackRide &&
				!loadingFallback
			) {
				try {
					setLoadingFallback(true);
					// Get the latest ride (first page, limit 1)
					const rides = await ApiService.getRides(1, 1);
					if (rides.length > 0) {
						// Get detailed info for the latest ride
						const rideDetail = await ApiService.getRideDetail(rides[0].id);
						setFallbackRide(rideDetail);
					}
				} catch (error) {
					console.error("Failed to fetch latest ride for fallback:", error);
				} finally {
					setLoadingFallback(false);
				}
			}
		};

		fetchLatestRideForFallback();
	}, [currentLocation, rideState.currentRide, fallbackRide, loadingFallback]);

	// Get fallback location from server's latest ride if no current location from websocket
	const getFallbackLocation = (): Position | null => {
		if (currentLocation) return currentLocation;

		// Try to get location from current ride's latest position (from websocket)
		if (rideState.currentRide && rideState.ridePositions.length > 0) {
			return rideState.ridePositions[rideState.ridePositions.length - 1];
		}

		// Try to get location from latest ride fetched from server
		if (fallbackRide && fallbackRide.positions.length > 0) {
			return fallbackRide.positions[fallbackRide.positions.length - 1];
		}

		// No fallback available
		return null;
	};

	const effectiveLocation = getFallbackLocation();

	// Get only the latest point for live view
	const getLatestPoint = (): LatLng[] => {
		// Priority 1: Current ride from websocket (latest position)
		if (rideState.currentRide && rideState.ridePositions.length > 0) {
			const latestPosition =
				rideState.ridePositions[rideState.ridePositions.length - 1];
			return [
				{
					lat: latestPosition.latitude,
					lng: latestPosition.longitude,
					speed_knots: latestPosition.speed_knots,
				},
			];
		}

		// Priority 2: Fallback ride from server (latest position)
		if (fallbackRide && fallbackRide.positions.length > 0) {
			const latestPosition =
				fallbackRide.positions[fallbackRide.positions.length - 1];
			return [
				{
					lat: latestPosition.latitude,
					lng: latestPosition.longitude,
					speed_knots: latestPosition.speed_knots,
				},
			];
		}

		// Priority 3: Just current location if available
		if (effectiveLocation) {
			return [
				{
					lat: effectiveLocation.latitude,
					lng: effectiveLocation.longitude,
					speed_knots: effectiveLocation.speed_knots,
				},
			];
		}

		return [];
	};

	const latLngList = getLatestPoint();

	const currentPosition = effectiveLocation
		? {
				lat: effectiveLocation.latitude,
				lng: effectiveLocation.longitude,
			}
		: undefined;

	return (
		<div className={styles.liveRoute}>
			<MapComponent latLngList={latLngList} currentPosition={currentPosition} />
			{/* {currentLocation && (
				<div className={styles.currentInfo}>
					<h3>Current Location</h3>
					{currentLocation.speed_knots && (
						<div>Speed: {formatSpeedMph(currentLocation.speed_knots)} mph</div>
					)}
					<div>
						Time: {new Date(currentLocation.timestamp).toLocaleTimeString()}
					</div>
				</div>
			)}
			{rideState.currentRide && (
				<div className={styles.rideInfo}>
					<h3>{rideState.currentRide.name}</h3>
					<div>
						Started:{" "}
						{new Date(rideState.currentRide.start_time).toLocaleTimeString()}
					</div>
					<div>Points: {rideState.ridePositions.length}</div>
					{!rideState.currentRide.end_time && (
						<div className={styles.ongoing}>🔴 Live</div>
					)}
				</div>
			)} */}
		</div>
	);
}
