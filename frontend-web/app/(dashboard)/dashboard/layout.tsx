"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { useAuth, UserButton } from "@clerk/nextjs";
import { api, User } from "../../lib/api";
import { 
	LayoutDashboard, 
	Users, 
	Calendar, 
	ShieldAlert, 
	Menu, 
	X, 
	Building, 
	LogOut,
	ShieldAlert as ShieldIcon
} from "lucide-react";

export default function DashboardLayout({
	children,
}: {
	children: React.ReactNode;
}) {
	const { getToken, isLoaded, isSignedIn } = useAuth();
	const router = useRouter();
	const pathname = usePathname();
	const [user, setUser] = useState<User | null>(null);
	const [loading, setLoading] = useState(true);
	const [sidebarOpen, setSidebarOpen] = useState(false);

	useEffect(() => {
		if (isLoaded && !isSignedIn) {
			router.push("/");
			return;
		}

		async function fetchUser() {
			try {
				const token = await getToken();
				if (token) {
					// We call /users/sync or /users/me. Let's try /users/me first.
					// If it fails with 404, we sync first using a mock placeholder email from Clerk.
					let profile: User;
					try {
						profile = await api.getMe(token);
					} catch (err) {
						// Fallback: sync user details
						profile = await api.syncUser(token, "console-user@flock.local", "Console User", "");
					}
					setUser(profile);
				}
			} catch (err) {
				console.error("Failed to load user profile:", err);
			} finally {
				setLoading(false);
			}
		}

		if (isSignedIn) {
			fetchUser();
		}
	}, [isLoaded, isSignedIn, getToken, router]);

	if (loading) {
		return (
			<div className="min-h-screen bg-zinc-950 flex flex-col items-center justify-center gap-4">
				<div className="h-10 w-10 border-4 border-violet-600 border-t-transparent rounded-full animate-spin" />
				<p className="text-sm text-zinc-400 font-medium">Authenticating & loading profile...</p>
			</div>
		);
	}

	if (!user) {
		return (
			<div className="min-h-screen bg-zinc-950 flex flex-col items-center justify-center p-6 text-center">
				<div className="h-12 w-12 rounded-xl bg-red-500/10 border border-red-500/25 flex items-center justify-center text-red-400 mb-4">
					<ShieldIcon className="h-6 w-6" />
				</div>
				<h1 className="text-xl font-bold mb-2">Access Denied</h1>
				<p className="text-sm text-zinc-400 max-w-sm mb-6">
					Unable to retrieve your user account profile. Please ensure you are authenticated properly.
				</p>
				<Link href="/" className="px-4 py-2 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 rounded-lg text-sm text-zinc-200">
					Back to Home
				</Link>
			</div>
		);
	}

	// RBAC Guards:
	const isOrganizer = user.role === "organizer" || user.role === "admin";
	const isAdmin = user.role === "admin";

	const navigation = [
		{ name: "Console Overview", href: "/dashboard", icon: LayoutDashboard, show: true },
		{ name: "Organizer Panel", href: "/dashboard/organizer", icon: Calendar, show: isOrganizer },
		{ name: "Admin Control", href: "/dashboard/admin", icon: ShieldAlert, show: isAdmin },
	];

	return (
		<div className="min-h-screen flex bg-zinc-950 text-zinc-50 font-sans selection:bg-violet-500/35 selection:text-white">
			{/* Sidebar for Desktop */}
			<aside className="hidden md:flex md:w-64 md:flex-col md:fixed md:inset-y-0 border-r border-zinc-900 bg-zinc-950/40 backdrop-blur-md z-20">
				<div className="flex flex-col flex-1 min-h-0">
					{/* Logo Header */}
					<div className="flex items-center h-16 flex-shrink-0 px-6 border-b border-zinc-900 gap-2">
						<div className="h-8 w-8 rounded-lg bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center font-bold text-white shadow-lg shadow-violet-500/20">
							F
						</div>
						<span className="text-lg font-bold tracking-tight bg-gradient-to-r from-white to-zinc-400 bg-clip-text text-transparent">
							Flock Console
						</span>
					</div>

					{/* Navigation Links */}
					<nav className="flex-1 px-4 py-6 space-y-1 overflow-y-auto">
						{navigation.map((item) => {
							if (!item.show) return null;
							const active = pathname === item.href;
							return (
								<Link
									key={item.name}
									href={item.href}
									className={`flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-xl transition-all ${
										active
											? "bg-violet-600/10 text-violet-400 border border-violet-500/20 shadow-md shadow-violet-500/5"
											: "text-zinc-400 hover:text-white hover:bg-zinc-900/50 border border-transparent"
									}`}
								>
									<item.icon className={`h-5 w-5 ${active ? "text-violet-400" : "text-zinc-500"}`} />
									{item.name}
								</Link>
							);
						})}
					</nav>

					{/* User Profile Card */}
					<div className="p-4 border-t border-zinc-900 flex items-center justify-between gap-2 bg-zinc-950/60">
						<div className="flex items-center gap-3 min-w-0">
							<UserButton />
							<div className="text-left min-w-0">
								<p className="text-xs font-semibold text-zinc-200 truncate max-w-[120px]">
									{user.full_name || "Console User"}
								</p>
								<p className="text-[10px] text-zinc-500 capitalize tracking-wider">
									{user.role}
								</p>
							</div>
						</div>
						<span className={`text-[10px] font-bold px-2 py-0.5 rounded-full uppercase ${
							user.role === "admin"
								? "bg-red-500/10 text-red-400 border border-red-500/20"
								: user.role === "organizer"
								? "bg-violet-500/10 text-violet-400 border border-violet-500/20"
								: "bg-zinc-800 text-zinc-400"
						}`}>
							{user.role}
						</span>
					</div>
				</div>
			</aside>

			{/* Sidebar for Mobile Drawer */}
			{sidebarOpen && (
				<div className="relative z-50 md:hidden">
					<div className="fixed inset-0 bg-black/60 backdrop-blur-sm" onClick={() => setSidebarOpen(false)} />
					<div className="fixed inset-y-0 left-0 w-64 bg-zinc-950 border-r border-zinc-900 flex flex-col">
						<div className="flex items-center h-16 px-6 justify-between border-b border-zinc-900">
							<div className="flex items-center gap-2">
								<div className="h-8 w-8 rounded-lg bg-gradient-to-tr from-violet-600 to-indigo-600 flex items-center justify-center font-bold text-white">
									F
								</div>
								<span className="text-lg font-bold">Flock</span>
							</div>
							<button className="p-1 hover:bg-zinc-900 rounded-lg" onClick={() => setSidebarOpen(false)}>
								<X className="h-5 w-5" />
							</button>
						</div>
						<nav className="flex-1 px-4 py-6 space-y-1">
							{navigation.map((item) => {
								if (!item.show) return null;
								const active = pathname === item.href;
								return (
									<Link
										key={item.name}
										href={item.href}
										onClick={() => setSidebarOpen(false)}
										className={`flex items-center gap-3 px-4 py-3 text-sm font-medium rounded-xl transition-all ${
											active
												? "bg-violet-600/10 text-violet-400 border border-violet-500/20"
												: "text-zinc-400 hover:text-white hover:bg-zinc-900/50"
										}`}
									>
										<item.icon className="h-5 w-5" />
										{item.name}
									</Link>
								);
							})}
						</nav>
						<div className="p-4 border-t border-zinc-900 flex items-center justify-between">
							<div className="flex items-center gap-3">
								<UserButton />
								<div className="text-left">
									<p className="text-xs font-semibold text-zinc-200">{user.full_name || "Console User"}</p>
									<p className="text-[10px] text-zinc-500 capitalize">{user.role}</p>
								</div>
							</div>
						</div>
					</div>
				</div>
			)}

			{/* Main Layout Area */}
			<div className="flex-1 flex flex-col md:pl-64">
				{/* Top Mobile Navbar */}
				<header className="sticky top-0 z-10 h-16 border-b border-zinc-900 bg-zinc-950/70 backdrop-blur-md flex items-center justify-between px-6 md:px-8">
					<button
						className="p-1.5 hover:bg-zinc-900 rounded-lg md:hidden"
						onClick={() => setSidebarOpen(true)}
					>
						<Menu className="h-5 w-5" />
					</button>

					<div className="flex items-center gap-4 ml-auto">
						<span className="hidden sm:inline-flex text-xs px-3 py-1 rounded-full border border-zinc-800 bg-zinc-900/40 text-zinc-400 font-medium">
							Server: Active
						</span>
						<div className="md:hidden">
							<UserButton />
						</div>
					</div>
				</header>

				{/* Dashboard Content Container */}
				<main className="flex-1 p-6 md:p-8 overflow-y-auto w-full max-w-7xl mx-auto">
					{children}
				</main>
			</div>
		</div>
	);
}
