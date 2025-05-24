import { useContext } from "react";
import type { LatLngContextType } from "../contexts/LatLngContext";
import { LatLngContext } from "../providers/LatLngProvider";

export function useLatLng(): LatLngContextType {
	const context = useContext(LatLngContext);
	if (context === undefined) {
		throw new Error("useLatLng must be used within a LatLngProvider");
	}
	return context;
}
