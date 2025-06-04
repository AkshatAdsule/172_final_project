import { Outlet, createFileRoute } from "@tanstack/react-router";
import { APIProvider } from "@vis.gl/react-google-maps";
import Sidebar from "../components/sidebar";
import { LatLngProvider } from "../providers/LatLngProvider";
import styles from "./rides.module.css";

export const Route = createFileRoute("/rides")({
	component: RouteComponent,
});

function RouteComponent() {
	const wsUrl = import.meta.env.VITE_API_BASE
		? `ws${import.meta.env.VITE_API_BASE}/ws`
		: "ws://localhost:8080/ws";
	const gmapsApiKey = import.meta.env.VITE_GOOGLE_MAPS_API_KEY;

	return (
		<div className={styles.container}>
			<div className={styles.sidebar}>
				<Sidebar />
			</div>
			<main className={styles.main}>
				<APIProvider apiKey={gmapsApiKey}>
					<LatLngProvider url={wsUrl}>
						<Outlet />
					</LatLngProvider>
				</APIProvider>
			</main>
		</div>
	);
}
