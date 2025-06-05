import { createFileRoute } from "@tanstack/react-router";
import { MapComponent } from "../../components/map/MapComponent";
import { useLatLng } from "../../hooks/useLatLng";
// import { formatSpeedMph } from "../../utils/speed";
import styles from "./styles/live.module.css";
import type { LatLng } from "../../types";

export const Route = createFileRoute("/rides/live")({
	component: RouteComponent,
});

function RouteComponent() {
	const { currentLocation, rideState } = useLatLng();

	const latLngList: LatLng[] = rideState.currentRide
		? rideState.ridePositions.map((p) => ({
				lat: p.latitude,
				lng: p.longitude,
				speed_knots: p.speed_knots,
			}))
		: currentLocation
			? [
					{
						lat: currentLocation.latitude,
						lng: currentLocation.longitude,
						speed_knots: currentLocation.speed_knots,
					},
				]
			: [];

	const currentPosition = currentLocation
		? {
				lat: currentLocation.latitude,
				lng: currentLocation.longitude,
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
