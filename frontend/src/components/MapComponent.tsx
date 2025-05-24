import { Map, useMap } from "@vis.gl/react-google-maps";
import { useEffect } from "react";
import { Polyline } from "../routes/rides/components/-polyline"; // Adjusted path
import type { LatLng } from "../types"; // Assuming LatLng type is defined in src/types

interface MapComponentProps {
	latLngList: LatLng[];
	defaultCenter?: { lat: number; lng: number };
	defaultZoom?: number;
}

export function MapComponent({
	latLngList,
	defaultCenter = { lat: 37.774929, lng: -122.419418 }, // Default to SF
	defaultZoom = 17,
}: MapComponentProps) {
	const map = useMap();

	// Setup polyline and calculate bounds
	useEffect(() => {
		if (!map) {
			if (map === undefined) {
				console.error(
					"MapComponent's internal map instance is not available. Ensure it is rendered within a valid APIProvider context.",
				);
			}
			return;
		}

		if (latLngList.length > 0) {
			const bounds = new google.maps.LatLngBounds();
			latLngList.forEach((latLng) => {
				bounds.extend(new google.maps.LatLng(latLng.lat, latLng.lng));
			});
			map.fitBounds(bounds);
		}
	}, [map, latLngList]);

	return (
		<Map defaultCenter={defaultCenter} defaultZoom={defaultZoom}>
			<Polyline path={latLngList} />
		</Map>
	);
}
