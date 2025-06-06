import { Link, createFileRoute } from "@tanstack/react-router";
import styles from "./styles/index.module.css";

export const Route = createFileRoute("/rides/")({
	component: RouteComponent,
});

function RouteComponent() {
	return (
		<div className={styles.emptyState}>
			<span>
				View <Link to="/rides/live">live location </Link>
				or select a ride from your history.
			</span>
		</div>
	);
}
