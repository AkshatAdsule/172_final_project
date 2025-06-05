import { Link, createFileRoute } from "@tanstack/react-router";
import styles from "./index.module.css";
import { Button } from "../components/button";

export const Route = createFileRoute("/")({
	component: Index,
});

function Index() {
	return (
		<>
			<div className={styles.hero}>
				<div className={styles.desc}>
					<h1>Your Bike's Ultimate Companion</h1>
					<p>
						Track rides, detect crashes, and prevent theft with the most
						advanced bike security system.
					</p>
					<div className={styles.ctaButtons}>
						<Button asChild variant="default" size="lg">
							<Link to="/rides/live">Start Live Tracking</Link>
						</Button>
						<Button asChild variant="outline" size="lg">
							<Link to="/rides">View Dashboard</Link>
						</Button>
					</div>
				</div>
				<img src="/b3_logo.png" alt="B3 Logo" />
			</div>

			<section className={styles.features}>
				<div className={styles.featuresHeader}>
					<h2>Core Features</h2>
					<p>Everything you need to keep your bike safe and track your rides</p>
				</div>

				<div className={styles.featureGrid}>
					<div className={styles.featureCard}>
						<div className={styles.featureIcon}>📍</div>
						<h3>Automatic Ride Tracking</h3>
						<p>
							Seamlessly record every ride with precise GPS tracking. View
							detailed routes, speed, distance, and elevation data in your
							dashboard.
						</p>
						<ul className={styles.featureList}>
							<li>Real-time GPS tracking</li>
							<li>Detailed route history</li>
							<li>Speed and distance metrics</li>
						</ul>
					</div>

					<div className={styles.featureCard}>
						<div className={styles.featureIcon}>🚨</div>
						<h3>Crash Detection</h3>
						<p>
							Intelligent impact sensors instantly detect crashes and alert your
							trusted contacts with exact location, time, and impact force data.
						</p>
						<ul className={styles.featureList}>
							<li>Instant crash detection</li>
							<li>Emergency contact alerts</li>
							<li>Impact force measurement</li>
							<li>Precise location tracking</li>
						</ul>
					</div>

					<div className={styles.featureCard}>
						<div className={styles.featureIcon}>🔒</div>
						<h3>Smart Bike Lock</h3>
						<p>
							Lock your bike remotely through the app. Get instant notifications
							if your bike moves while secured, preventing theft before it
							happens.
						</p>
						<ul className={styles.featureList}>
							<li>Remote locking capability</li>
							<li>Movement detection</li>
							<li>Instant theft alerts</li>
							<li>Live location tracking</li>
						</ul>
					</div>
				</div>
			</section>

			<section className={styles.architecture}>
				<div className={styles.architectureHeader}>
					<h2>System Architecture</h2>
					<p>
						Understanding how B3 components work together to keep your bike safe
					</p>
				</div>
				<div className={styles.architectureContent}>
					<div className={styles.architectureDiagram}>
						<img src="/b3_arch.svg" alt="B3 System Architecture Diagram" />
					</div>
					<div className={styles.architectureDescription}>
						<h3>How It Works</h3>
						<p>
							The B3 system consists of multiple interconnected components that
							work together to provide comprehensive bike security and tracking:
						</p>
						<ul className={styles.architectureList}>
							<li>On-bike sensors continuously monitor movement and impact</li>
							<li>GPS module provides real-time location tracking</li>
							<li>Location is constantly sent to AWS IoT Core (~1 update/s)</li>
							<li>Onboard accelerometer detects crashes</li>
							<li>Cloud backend processes and stores all data</li>
							<li>
								Backend dispatches theft and crash notifications via Amazon
								Simple Notification Service (SNS)
							</li>
							<li>Backend also detects and records rides automatically</li>
							<li>
								Frontend connects to the backend via HTTP and WebSocket for
								real-time updates
							</li>
						</ul>
					</div>
				</div>
			</section>

			<section className={styles.stateDiagrams}>
				<div className={styles.stateDiagramsHeader}>
					<h2>System State Diagrams</h2>
					<p>
						Understanding the behavior of our server and microcontroller
						components
					</p>
				</div>
				<div className={styles.stateDiagramsContent}>
					<div className={styles.stateDiagramCard}>
						<h3>Server State Machine</h3>
						<div className={styles.diagramContainer}>
							<img src="/b3_server_state.png" alt="Server State Diagram" />
						</div>
						<p>
							The server manages three primary states: Idle (awaiting ride
							initiation), Tracking (actively recording an ongoing ride), and
							Paused (ride ongoing but device stationary). The system begins in
							the Idle state and transitions to Tracking when it receives valid
							GPS data indicating the device has moved beyond a start threshold,
							triggering ride creation in the database and various broadcast
							notifications to the frontend. While in the Tracking state, the
							server continuously processes GPS updates, adding position data to
							the ride and broadcasting updates as long as the device remains
							moving; however, if the device becomes stationary (falls below the
							static threshold), it transitions to the Paused state while
							maintaining the active ride. From the Paused state, the system can
							either resume tracking if movement is detected again or
							automatically end the ride if static timeout or general inactivity
							timeout conditions are met. Throughout all states, the server
							handles crash detection by sending SNS notifications and includes
							robust timeout mechanisms to automatically end rides that become
							inactive, ensuring proper cleanup of stale ride sessions and
							maintaining system reliability through periodic inactivity checks.
						</p>
					</div>
					<div className={styles.stateDiagramCard}>
						<h3>MCU State Machine</h3>
						<div className={styles.diagramContainer}>
							<img src="/b3_3200_state.svg" alt="MCU State Diagram" />
						</div>
						<p>
							The microcontroller starts with a Startup Phase where it boots up,
							connects to Wi-Fi, establishes a TLS connection, and acquires a
							GPS fix. Following this, it enters the Main Loop Phase,
							continuously gathering GPS data and reading the onboard
							accelerometer. If the accelerometer detects an impact, the MCU
							updates a designated "Crash Shadow" on AWS and shows a crash
							screen; otherwise, it proceeds to update the standard AWS Shadow
							with current data. A successful AWS Shadow update results in an
							update to an OLED display. However, if the shadow update fails,
							the MCU transitions to the Network Recovery Phase. In this phase,
							it first checks its connection status: if connected but issues
							persist (like the failed shadow update), it attempts to re-setup
							the TLS connection before trying to resume main loop operations.
							If not connected, it disconnects from the current access point,
							attempts to reconnect to the hotspot AP, and then re-establishes
							the TLS connection, aiming to return to a stable point in the main
							loop.
						</p>
					</div>
				</div>
			</section>

			<section className={styles.demoSection}>
				<div className={styles.demoHeader}>
					<h2>See B3 in Action</h2>
					<p>
						Watch a live demo of the B3 system tracking rides and detecting
						crashes.
					</p>
				</div>
				<div className={styles.demoVideoContainer}>
					<video controls poster="/poster.png" className={styles.demoVideo}>
						<source src="/demo_optimized.mp4" type="video/mp4" />
						Your browser does not support the video tag.
					</video>
				</div>
			</section>
		</>
	);
}
