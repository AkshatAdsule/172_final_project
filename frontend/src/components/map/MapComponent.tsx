import { Map, useMap } from "@vis.gl/react-google-maps";
import { useEffect, useState, useCallback } from "react";
import { Polyline } from "./polyline";
import { SpeedPolyline } from "./SpeedPolyline";
import { SpeedLegend } from "./SpeedLegend";
import type { LatLng } from "../../types";
import { DARK_STYLE } from "./dark-style";

interface MapComponentProps {
	latLngList?: LatLng[];
	defaultCenter?: { lat: number; lng: number };
	defaultZoom?: number;
}

// Helper function to check if path has speed data
function hasSpeedData(path: LatLng[]): boolean {
	return path.some((point) => point.speed_knots !== undefined);
}

export function MapComponent({
	latLngList = [],
	defaultCenter = { lat: 37.774929, lng: -122.419418 }, // Default to SF
	defaultZoom = 19,
}: MapComponentProps) {
	const map = useMap();
	const [speedRange, setSpeedRange] = useState<{
		min: number;
		max: number;
	} | null>(null);

	// Determine whether to use speed polyline
	const useSpeedPolyline = hasSpeedData(latLngList);

	// Handle speed range updates from SpeedPolyline - wrapped in useCallback to prevent infinite loops
	const handleSpeedRangeChange = useCallback(
		(minSpeed: number, maxSpeed: number) => {
			setSpeedRange({ min: minSpeed, max: maxSpeed });
		},
		[],
	);

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

			if (map.getZoom()! > defaultZoom) {
				map.setZoom(defaultZoom);
			}
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
		<div style={{ position: "relative", width: "100%", height: "100%" }}>
			<Map
				styles={DARK_STYLE}
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
				{useSpeedPolyline ? (
					<SpeedPolyline
						path={latLngList}
						onSpeedRangeChange={handleSpeedRangeChange}
					/>
				) : (
					<Polyline path={latLngList} strokeColor={"#ffffffde"} />
				)}
			</Map>
			{useSpeedPolyline && speedRange && speedRange.min !== speedRange.max && (
				<SpeedLegend minSpeed={speedRange.min} maxSpeed={speedRange.max} />
			)}
		</div>
	);
}
