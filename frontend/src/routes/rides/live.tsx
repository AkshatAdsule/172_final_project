import { createFileRoute } from "@tanstack/react-router";
import { Map, useMap } from "@vis.gl/react-google-maps";
import { useLatLng } from "../../hooks/useLatLng";
import { useEffect } from "react";
import { Polyline } from "./_components/-polyline";
import "./styles/live.css";

export const Route = createFileRoute("/rides/live")({
	component: RouteComponent,
});

function RouteComponent() {
	const { latLngList } = useLatLng();
	const map = useMap();

	// setup polyline
	useEffect(() => {
		if (!map) {
			if (map === undefined)
				console.error("<Polyline> has to be inside a Map component.");

			return;
		}
	}, [map]);

	// calculate bounds to fit all lat lngs
	useEffect(() => {
		if (map && latLngList.length > 0) {
			const bounds = new google.maps.LatLngBounds();
			latLngList.forEach((latLng) => {
				bounds.extend(new google.maps.LatLng(latLng.lat, latLng.lng));
			});
			map.fitBounds(bounds);
		}
	}, [map, latLngList]);

	return (
		<>
			<Map
				// for now center in sf
				defaultCenter={{
					lat: 37.774929,
					lng: -122.419418,
				}}
				defaultZoom={17}
			>
				<Polyline path={latLngList} />
			</Map>
		</>
	);
}
