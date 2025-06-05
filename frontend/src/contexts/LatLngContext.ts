import { createContext } from "react";
import type { ReadyStateString } from "../hooks/websocket";
import type { Position, RideSummary } from "../types";

interface RideState {
	currentRide: RideSummary | null;
	ridePositions: Position[];
}

export interface LatLngContextType {
	readyState: ReadyStateString;
	sendMessage: (
		data: string | ArrayBufferLike | Blob | ArrayBufferView,
	) => void;
	currentLocation: Position | null;
	rideState: RideState;
}

export const LatLngContext = createContext<LatLngContextType | undefined>(
	undefined,
);
