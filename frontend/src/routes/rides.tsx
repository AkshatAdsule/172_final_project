import { Outlet, createFileRoute } from "@tanstack/react-router";
import { APIProvider } from "@vis.gl/react-google-maps";
import Sidebar from "../components/sidebar";
import { LatLngProvider } from "../providers/LatLngProvider";
import styles from "./rides.module.css";

export const Route = createFileRoute("/rides")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<div className={styles.container}>
			<main className={styles.main}>
				<APIProvider apiKey={import.meta.env.VITE_GOOGLE_MAPS_API_KEY}>
					<LatLngProvider url="ws://localhost:8080/ws">
						<Outlet />
					</LatLngProvider>
				</APIProvider>
			</main>
			<div className={styles.sidebar}>
				<Sidebar />
			</div>
		</div>
	);
}
