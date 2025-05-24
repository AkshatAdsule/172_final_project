import { createContext } from "react";
import type { ReadyStateString } from "../hooks/websocket";
import type { LatLng } from "../types";

export interface LatLngContextType {
	latLngList: LatLng[];
	readyState: ReadyStateString;
	sendMessage: (
		data: string | ArrayBufferLike | Blob | ArrayBufferView,
	) => void;
}

export const LatLngContext = createContext<LatLngContextType | undefined>(
	undefined,
);
