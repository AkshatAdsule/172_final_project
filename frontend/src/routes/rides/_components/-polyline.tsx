import {
	forwardRef,
	useContext,
	useEffect,
	useImperativeHandle,
	useMemo,
	useRef,
} from "react";

import { GoogleMapsContext, useMapsLibrary } from "@vis.gl/react-google-maps";

import type { Ref } from "react";

type PolylineCustomProps = {
	/**
	 * this is an encoded string for the path, will be decoded and used as a path
	 */
	path: { lat: number; lng: number }[];
};

export type PolylineProps = google.maps.PolylineOptions & PolylineCustomProps;

export type PolylineRef = Ref<google.maps.Polyline | null>;

function usePolyline(props: PolylineProps) {
	const { path, ...polylineOptions } = props;

	const geometryLibrary = useMapsLibrary("geometry");

	const polyline = useRef(new google.maps.Polyline()).current;
	// update PolylineOptions (note the dependencies aren't properly checked
	// here, we just assume that setOptions is smart enough to not waste a
	// lot of time updating values that didn't change)
	useMemo(() => {
		polyline.setOptions(polylineOptions);
	}, [polyline, polylineOptions]);

	const map = useContext(GoogleMapsContext)?.map;

	// update the path with the encodedPath
	useMemo(() => {
		// if (!encodedPath || !geometryLibrary) return;
		// const path = geometryLibrary.encoding.decodePath(encodedPath);
		// polyline.setPath(path);
		polyline.setPath(
			path.map((point) => new google.maps.LatLng(point.lat, point.lng)),
		);
	}, [polyline, path, geometryLibrary]);

	// create polyline instance and add to the map once the map is available
	useEffect(() => {
		if (!map) {
			if (map === undefined)
				console.error("<Polyline> has to be inside a Map component.");

			return;
		}

		polyline.setMap(map);

		return () => {
			polyline.setMap(null);
		};
	}, [map]);

	return polyline;
}

/**
 * Component to render a polyline on a map
 */
export const Polyline = forwardRef((props: PolylineProps, ref: PolylineRef) => {
	const polyline = usePolyline(props);

	useImperativeHandle(ref, () => polyline, []);

	return null;
});
