import { createFileRoute } from "@tanstack/react-router";
import { MapComponent } from "../../components/map/MapComponent";
import { useLatLng } from "../../hooks/useLatLng";
import styles from "./styles/live.module.css";

export const Route = createFileRoute("/rides/live")({
	component: RouteComponent,
});

function RouteComponent() {
	const { latLngList, currentLocation, rideState } = useLatLng();

	return (
		<div className={styles.liveRoute}>
			<MapComponent latLngList={latLngList} />
			{currentLocation && (
				<div className={styles.currentInfo}>
					<h3>Current Location</h3>
					<div>Lat: {currentLocation.latitude.toFixed(6)}</div>
					<div>Lng: {currentLocation.longitude.toFixed(6)}</div>
					{currentLocation.speed_knots && (
						<div>Speed: {currentLocation.speed_knots.toFixed(1)} knots</div>
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
			)}
		</div>
	);
}
