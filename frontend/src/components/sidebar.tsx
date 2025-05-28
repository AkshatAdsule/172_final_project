import { Link } from "@tanstack/react-router";
import { useRides } from "../hooks/useRides";
import styles from "./sidebar.module.css";

export default function Sidebar() {
	const { rides, loading, error } = useRides();

	const formatDate = (dateString: string) => {
		return new Date(dateString).toLocaleDateString();
	};

	const formatTime = (dateString: string) => {
		return new Date(dateString).toLocaleTimeString([], {
			hour: '2-digit',
			minute: '2-digit'
		});
	};

	return (
		<nav className={styles.nav}>
			<div className={styles.brand}>
				<h1>
					B<sup>3</sup>
				</h1>
			</div>
			<hr />
			<div className={styles.rides}>
				<Link to="/rides/live">
					<div>Live Track</div>
				</Link>
				<hr />

				{loading && <div className={styles.loading}>Loading rides...</div>}
				{error && <div className={styles.error}>Error: {error}</div>}

				{!loading && !error && rides.length === 0 && (
					<div className={styles.empty}>No rides found</div>
				)}

				{!loading && !error && rides.map((ride) => (
					<Link key={ride.id} to={"/rides/$rideId"} params={{ rideId: ride.id.toString() }}>
						<div className={styles.ride}>
							<span className={styles.name}>{ride.name}</span>
							<span className={styles.date}>
								{formatDate(ride.start_time)}
							</span>
							<span className={styles.time}>
								{formatTime(ride.start_time)}
								{ride.end_time && ` - ${formatTime(ride.end_time)}`}
							</span>
						</div>
					</Link>
				))}
			</div>
		</nav>
	);
}
