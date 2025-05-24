import { createFileRoute } from "@tanstack/react-router";
import { useLatLng } from "../../hooks/useLatLng";
import "./styles/live.css";
import { MapComponent } from "../../components/MapComponent"; // Import the new component

export const Route = createFileRoute("/rides/live")({
	component: RouteComponent,
});

function RouteComponent() {
	const { latLngList } = useLatLng();

	return (
		<div className="live-route">
			<MapComponent latLngList={latLngList} />
		</div>
	);
}
