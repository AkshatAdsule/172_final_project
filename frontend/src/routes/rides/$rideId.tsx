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
    const latLngList: LatLng[] = rideDetail.positions.map(position => ({
        lat: position.latitude,
        lng: position.longitude
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
                </div>
            </div>
        </div>
    );
} 