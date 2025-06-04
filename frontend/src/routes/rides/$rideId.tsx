import { createFileRoute } from "@tanstack/react-router";
import { MapComponent } from "../../components/map/MapComponent";
import { useRideDetail } from "../../hooks/useRides";
import type { LatLng } from "../../types";
import styles from "./styles/rideDetail.module.css";

export const Route = createFileRoute("/rides/$rideId")({
	component: RouteComponent,
});

function RouteComponent() {
	const { rideId } = Route.useParams();
	const { rideDetail, loading, error } = useRideDetail(Number(rideId));

	if (loading) {
		return (
			<div className={styles.loading}>
				<div>Loading ride details...</div>
			</div>
		);
	}

	if (error) {
		return (
			<div className={styles.error}>
				<div>Error loading ride: {error}</div>
			</div>
		);
	}

	if (!rideDetail) {
		return (
			<div className={styles.notFound}>
				<div>Ride not found</div>
			</div>
		);
	}

	// Convert positions to LatLng format for the map
	const latLngList: LatLng[] = rideDetail.positions.map((position) => ({
		lat: position.latitude,
		lng: position.longitude,
		speed_knots: position.speed_knots,
	}));

	const formatDateTime = (dateString: string) => {
		return new Date(dateString).toLocaleString();
	};

	const calculateDuration = () => {
		if (!rideDetail.end_time) return "Ongoing";

		const start = new Date(rideDetail.start_time);
		const end = new Date(rideDetail.end_time);
		const durationMs = end.getTime() - start.getTime();
		const minutes = Math.floor(durationMs / 60000);
		const hours = Math.floor(minutes / 60);

		if (hours > 0) {
			return `${hours}h ${minutes % 60}m`;
		}
		return `${minutes}m`;
	};

	const calculateStats = () => {
		const speeds = rideDetail.positions
			.map((p) => p.speed_knots)
			.filter((speed): speed is number => speed !== undefined);

		if (speeds.length === 0) return null;

		const avgSpeed =
			speeds.reduce((sum, speed) => sum + speed, 0) / speeds.length;
		const maxSpeed = Math.max(...speeds);

		return { avgSpeed, maxSpeed };
	};

	const stats = calculateStats();

	return (
		<div className={styles.rideDetail}>
			<div className={styles.mapContainer}>
				<MapComponent latLngList={latLngList} />
			</div>
			<div className={styles.info}>
				<h2>{rideDetail.name}</h2>
				<div className={styles.details}>
					<div className={styles.detailItem}>
						<span className={styles.label}>Started:</span>
						<span>{formatDateTime(rideDetail.start_time)}</span>
					</div>
					{rideDetail.end_time && (
						<div className={styles.detailItem}>
							<span className={styles.label}>Ended:</span>
							<span>{formatDateTime(rideDetail.end_time)}</span>
						</div>
					)}
					<div className={styles.detailItem}>
						<span className={styles.label}>Duration:</span>
						<span>{calculateDuration()}</span>
					</div>
					<div className={styles.detailItem}>
						<span className={styles.label}>Points:</span>
						<span>{rideDetail.positions.length}</span>
					</div>
					{stats && (
						<>
							<div className={styles.detailItem}>
								<span className={styles.label}>Avg Speed:</span>
								<span>{stats.avgSpeed.toFixed(1)} knots</span>
							</div>
							<div className={styles.detailItem}>
								<span className={styles.label}>Max Speed:</span>
								<span>{stats.maxSpeed.toFixed(1)} knots</span>
							</div>
						</>
					)}
				</div>
			</div>
		</div>
	);
}
