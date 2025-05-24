import { createFileRoute } from "@tanstack/react-router";
import { useLatLng } from "../hooks/useLatLng";

export const Route = createFileRoute("/")({
	component: Index,
});

function Index() {
	const { latLngList, readyState } = useLatLng();
	return (
		<div className="p-2">
			<p>Current WebSocket ready state: {readyState}</p>

			<p>
				Most Recent Lat Long: ({latLngList[latLngList.length - 1]?.lat},{" "}
				{latLngList[latLngList.length - 1]?.lon})
			</p>
		</div>
	);
}
