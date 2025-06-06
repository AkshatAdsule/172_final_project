import {
	forwardRef,
	useContext,
	useEffect,
	useImperativeHandle,
	useRef,
} from "react";

import { GoogleMapsContext } from "@vis.gl/react-google-maps";

import type { Ref } from "react";

type PolylineCustomProps = {
	path: { lat: number; lng: number }[];
};

export type PolylineProps = google.maps.PolylineOptions & PolylineCustomProps;

export type PolylineRef = Ref<google.maps.Polyline | null>;

function usePolyline(props: PolylineProps) {
	const { path, ...polylineOptions } = props;

	const map = useContext(GoogleMapsContext)?.map;
	const polylineRef = useRef<google.maps.Polyline | null>(null);

	// Create polyline instance when both map and geometry library are ready
	useEffect(() => {
		if (!map) return;

		// Create the polyline instance if it doesn't exist
		if (!polylineRef.current) {
			polylineRef.current = new google.maps.Polyline();
		}

		// Set the polyline options
		polylineRef.current.setOptions(polylineOptions);

		// Add to map
		polylineRef.current.setMap(map);

		return () => {
			if (polylineRef.current) {
				polylineRef.current.setMap(null);
			}
		};
	}, [map, polylineOptions]);

	// Update the path when it changes
	useEffect(() => {
		if (polylineRef.current && path.length > 0) {
			polylineRef.current.setPath(
				path.map((point) => new google.maps.LatLng(point.lat, point.lng)),
			);
		}
	}, [path]);

	// Cleanup on unmount
	useEffect(() => {
		return () => {
			if (polylineRef.current) {
				polylineRef.current.setMap(null);
				polylineRef.current = null;
			}
		};
	}, []);

	return polylineRef.current;
}

/**
 * Component to render a polyline on a map
 */
export const Polyline = forwardRef<google.maps.Polyline | null, PolylineProps>(
	(props, ref) => {
		const polyline = usePolyline(props);

		// @ts-ignore - polyline can be null when Google Maps API is still loading
		useImperativeHandle(ref, () => polyline, [polyline]);

		return null;
	},
);
