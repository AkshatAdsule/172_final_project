import { createFileRoute } from "@tanstack/react-router";
import { MapComponent } from "../../components/map/MapComponent"; // Import the new component
import { useLatLng } from "../../hooks/useLatLng";
import styles from "./styles/live.module.css";

export const Route = createFileRoute("/rides/live")({
	component: RouteComponent,
});

function RouteComponent() {
	const { latLngList } = useLatLng();

	return (
		<div className={styles.liveRoute}>
			<MapComponent latLngList={latLngList} />
		</div>
	);
}
