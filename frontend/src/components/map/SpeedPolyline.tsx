import {
	forwardRef,
	useContext,
	useEffect,
	useImperativeHandle,
	useMemo,
	useRef,
} from "react";

import { GoogleMapsContext } from "@vis.gl/react-google-maps";

import type { Ref } from "react";
import type { LatLng } from "../../types";

type SpeedPolylineCustomProps = {
	path: LatLng[];
	strokeWeight?: number;
	strokeOpacity?: number;
	onSpeedRangeChange?: (minSpeed: number, maxSpeed: number) => void;
};

export type SpeedPolylineProps = SpeedPolylineCustomProps;

export type SpeedPolylineRef = Ref<google.maps.Polyline[] | null>;

// Color interpolation function from blue to red based on speed
function interpolateColor(
	speed: number,
	minSpeed: number,
	maxSpeed: number,
): string {
	// Clamp speed between min and max
	const normalizedSpeed = Math.max(
		0,
		Math.min(1, (speed - minSpeed) / (maxSpeed - minSpeed)),
	);

	// Blue RGB: (0, 123, 255)
	// Red RGB: (255, 82, 82)
	const r = Math.round(normalizedSpeed * 255 + (1 - normalizedSpeed) * 0);
	const g = Math.round(normalizedSpeed * 82 + (1 - normalizedSpeed) * 123);
	const b = Math.round(normalizedSpeed * 82 + (1 - normalizedSpeed) * 255);

	return `rgb(${r}, ${g}, ${b})`;
}

function useSpeedPolyline(props: SpeedPolylineProps) {
	const {
		path,
		strokeWeight = 4,
		strokeOpacity = 0.8,
		onSpeedRangeChange,
	} = props;
	const polylinesRef = useRef<google.maps.Polyline[]>([]);
	const map = useContext(GoogleMapsContext)?.map;

	// Calculate speed statistics for color interpolation
	const { minSpeed, maxSpeed } = useMemo(() => {
		const speeds = path
			.map((point) => point.speed_knots)
			.filter((speed): speed is number => speed !== undefined);

		if (speeds.length === 0) {
			return { minSpeed: 0, maxSpeed: 10 }; // Default range if no speed data
		}

		return {
			minSpeed: Math.min(...speeds),
			maxSpeed: Math.max(...speeds),
		};
	}, [path]);

	// Notify parent component of speed range changes
	useEffect(() => {
		if (onSpeedRangeChange) {
			onSpeedRangeChange(minSpeed, maxSpeed);
		}
	}, [minSpeed, maxSpeed, onSpeedRangeChange]);

	// Create and manage polyline segments
	useEffect(() => {
		if (!map) return;

		// Clear existing polylines
		polylinesRef.current.forEach((polyline) => {
			polyline.setMap(null);
		});
		polylinesRef.current = [];

		if (path.length < 2) return;

		// Create segments between consecutive points
		for (let i = 0; i < path.length - 1; i++) {
			const startPoint = path[i];
			const endPoint = path[i + 1];

			// Use the speed of the start point for the segment color
			// If no speed data, use a default blue color
			const speed = startPoint.speed_knots ?? 0;
			const color = interpolateColor(speed, minSpeed, maxSpeed);

			const polyline = new google.maps.Polyline({
				path: [
					new google.maps.LatLng(startPoint.lat, startPoint.lng),
					new google.maps.LatLng(endPoint.lat, endPoint.lng),
				],
				strokeColor: color,
				strokeWeight,
				strokeOpacity,
				geodesic: true,
			});

			// Add to map
			polyline.setMap(map);
			polylinesRef.current.push(polyline);
		}

		return () => {
			polylinesRef.current.forEach((polyline) => {
				polyline.setMap(null);
			});
		};
	}, [map, path, strokeWeight, strokeOpacity, minSpeed, maxSpeed]);

	// Cleanup on unmount
	useEffect(() => {
		return () => {
			polylinesRef.current.forEach((polyline) => {
				polyline.setMap(null);
			});
			polylinesRef.current = [];
		};
	}, []);

	return polylinesRef.current;
}

/**
 * Component to render speed-colored polyline segments on a map
 */
export const SpeedPolyline = forwardRef<google.maps.Polyline[] | null, SpeedPolylineProps>(
	(props, ref) => {
		const polylines = useSpeedPolyline(props);

		// @ts-ignore - polylines can be empty array when Google Maps API is still loading
		useImperativeHandle(ref, () => polylines, [polylines]);

		return null;
	},
);
