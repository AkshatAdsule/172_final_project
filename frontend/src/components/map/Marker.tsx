import {
	forwardRef,
	useContext,
	useEffect,
	useImperativeHandle,
	useRef,
} from "react";

import { GoogleMapsContext } from "@vis.gl/react-google-maps";

import type { Ref } from "react";

type MarkerCustomProps = {
	position: { lat: number; lng: number };
};

export type MarkerProps = google.maps.MarkerOptions & MarkerCustomProps;

export type MarkerRef = Ref<google.maps.Marker | null>;

function useMarker(props: MarkerProps) {
	const { position, ...markerOptions } = props;

	const map = useContext(GoogleMapsContext)?.map;
	const markerRef = useRef<google.maps.Marker | null>(null);

	// Create marker instance when map is ready
	useEffect(() => {
		if (!map) return;

		// Create the marker instance if it doesn't exist
		if (!markerRef.current) {
			markerRef.current = new google.maps.Marker();
		}

		// Set the marker options
		markerRef.current.setOptions({
			...markerOptions,
			position: new google.maps.LatLng(position.lat, position.lng),
		});

		// Add to map
		markerRef.current.setMap(map);

		return () => {
			if (markerRef.current) {
				markerRef.current.setMap(null);
			}
		};
	}, [map, markerOptions]);

	// Update the position when it changes
	useEffect(() => {
		if (markerRef.current) {
			markerRef.current.setPosition(
				new google.maps.LatLng(position.lat, position.lng),
			);
		}
	}, [position]);

	// Cleanup on unmount
	useEffect(() => {
		return () => {
			if (markerRef.current) {
				markerRef.current.setMap(null);
				markerRef.current = null;
			}
		};
	}, []);

	return markerRef.current;
}

/**
 * Component to render a marker on a map
 */
export const Marker = forwardRef<google.maps.Marker | null, MarkerProps>(
	(props, ref) => {
		const marker = useMarker(props);

		// @ts-ignore - marker can be null when Google Maps API is still loading
		useImperativeHandle(ref, () => marker, [marker]);

		return null;
	},
);
