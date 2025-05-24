import { Map, useMap } from "@vis.gl/react-google-maps";
import { useEffect } from "react";
import { Polyline } from "../routes/rides/components/-polyline";
import type { LatLng } from "../types";

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

			// Adjust bounds to account for the floating sidebar
			map.fitBounds(bounds, {
				left: 340, // Sidebar width + margin
				top: 40,
				right: 40,
				bottom: 40,
			});
		} else {
			// Center the map considering the sidebar offset
			const center = new google.maps.LatLng(
				defaultCenter.lat,
				defaultCenter.lng,
			);
			map.setCenter(center);
			map.setZoom(defaultZoom);

			// Pan slightly right to center in visible area
			map.panBy(-170, 0); // Half of sidebar width
		}
	}, [map, latLngList, defaultCenter, defaultZoom]);

	return (
		<Map
			defaultCenter={defaultCenter}
			defaultZoom={defaultZoom}
			cameraControl={false}
			disableDefaultUI={true}
			zoomControl={false}
			gestureHandling={"none"}
			disableDoubleClickZoom={true}
			scrollwheel={false}
			style={{ width: "100%", height: "100%" }}
		>
			<Polyline path={latLngList} />
		</Map>
	);
}
