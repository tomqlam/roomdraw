import "@fortawesome/fontawesome-free/css/all.min.css";
import { GoogleLogin } from "@react-oauth/google";
import "bulma/css/bulma.min.css";
import { jwtDecode } from "jwt-decode";
import React, { useCallback, useContext, useEffect, useState, lazy, Suspense, useMemo } from "react";
import Select from "react-select";
import { CSSTransition, TransitionGroup } from "react-transition-group";
import BlocklistManager from "./Admin/BlocklistManager";
import Navbar from "./components/Navbar";
import FloorGrid from "./components/FloorGrid";
import { MyContext } from "./context/MyContext";
import SearchPage from "./pages/Search/SearchPage";
import "./styles.css";

// Lazy load modals for better bundle splitting
const BumpFroshModal = lazy(() => import("./modals/BumpFroshModal"));
const BumpModal = lazy(() => import("./modals/BumpModal"));
const FAQModal = lazy(() => import("./modals/FAQModal"));
const SettingsModal = lazy(() => import("./modals/SettingsModal"));
const SuiteNoteModal = lazy(() => import("./modals/SuiteNoteModal"));
const UserSettingsModal = lazy(() => import("./modals/UserSettingsModal"));

// Extracted component to prevent recreation on every render
const FloorDisplay = ({ gridData, filterCondition, showFloorplans, activeTab }) => {
    return (
        <div className="column">
            <div style={showFloorplans ? { width: "100%" } : {}}>
                {gridData.map((dorm) => (
                    <div key={dorm.dormName} className={activeTab === dorm.dormName ? "" : "is-hidden"}>
                        {dorm.floors
                            .filter((floor) => filterCondition(floor.floorNumber))
                            .sort((a, b) => Number(a.floorNumber) - Number(b.floorNumber))
                            .map((floor, floorIndex) => (
                                <div key={floorIndex} className="floor-section">
                                    <div className="floor-header">
                                        <h2 className="subtitle mb-2">Floor {floor.floorNumber + 1}</h2>
                                        {floor.floorName && <p className="subtitle mb-4">{floor.floorName}</p>}
                                    </div>
                                    <FloorGrid gridData={floor} />
                                </div>
                            ))}
                    </div>
                ))}
            </div>
        </div>
    );
};

function App() {
    const {
        currPage,
        gridData,
        userMap,
        isModalOpen,
        dormMapping,
        selectedID,
        setSelectedID,
        rooms,
        isSuiteNoteModalOpen,
        credentials,
        setCredentials,
        activeTab,
        setActiveTab,
        isFroshModalOpen,
        getRoomUUIDFromUserID,
        setRefreshKey,
        handleErrorFromTokenExpiry,
        isSettingsModalOpen,
        showFloorplans,
        userID,
        setUserID,
        refreshKey,
        isUserSettingsModalOpen,
        setIsUserSettingsModalOpen,
        handleTakeMeThere,
    } = useContext(MyContext);

    const [notifications, setNotifications] = useState([]);
    const [isTransitioning, setIsTransitioning] = useState(false);
    const [, setSelectedUserData] = useState(null);
    const [selectedUserRoom, setSelectedUserRoom] = useState("Unselected"); // to show what room current selected user is in
    const [isInDorm, setIsInDorm] = useState(true);
    const [lastSelectedID, setLastSelectedID] = useState(null); // Store last valid selection
    const [isSearchFocused, setIsSearchFocused] = useState(false); // Track if search is focused
    const [showFAQModal, setShowFAQModal] = useState(false);

    const userOptions = useMemo(() => {
        if (!userMap) return [];
        return Object.keys(userMap)
            .sort((a, b) => {
                const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`;
                const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`;
                return nameA.localeCompare(nameB);
            })
            .filter((key) => Number(userMap[key].Year) !== 0)
            .map((key) => ({
                value: key,
                label: `${userMap[key].FirstName} ${userMap[key].LastName}`,
            }));
    }, [userMap]);

    // Validate selectedID when userMap loads
    useEffect(() => {
        if (userMap && selectedID) {
            // Check if the selectedID exists in userMap
            if (!userMap[selectedID]) {
                // Reset selectedID if it doesn't exist in userMap
                setSelectedID(null);
                setSelectedUserRoom("Unselected");
            }
        }
    }, [userMap, selectedID, setSelectedID]);

    useEffect(() => {
        const closedNotifications = JSON.parse(localStorage.getItem("closedNotifications")) || [];
        const newNotifications = [
            "Reminder: to show/hide floorplans, click Personalization in the top right corner and toggle the checkbox.",
            // 'Notification 2',
        ].filter((notification) => !closedNotifications.includes(notification));
        setNotifications(newNotifications);
    }, []);

    const handleCloseNotification = (notification) => {
        // Store the notification state in local storage when the user closes the notification
        const closedNotifications = JSON.parse(localStorage.getItem("closedNotifications")) || [];
        closedNotifications.push(notification);
        localStorage.setItem("closedNotifications", JSON.stringify(closedNotifications));
        setNotifications(notifications.filter((n) => n !== notification));
    };

    useEffect(() => {
        const fetchRoomInfo = async () => {
            // Skip if no selectedID
            if (!selectedID) return;

            const roomUUID = getRoomUUIDFromUserID(selectedID);
            // Skip if no roomUUID
            if (!roomUUID) return;

            try {
                const response = await fetch(`${process.env.REACT_APP_API_URL}/rooms/${roomUUID}`, {
                    headers: {
                        Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                    },
                });
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const roomData = await response.json();
                if (handleErrorFromTokenExpiry(roomData)) {
                    return;
                }

                // Check if the room has inherited priority
                if (roomData.PullPriority.inherited.valid) {
                    setIsInDorm(roomData.PullPriority.inherited.hasInDorm);
                } else {
                    setIsInDorm(roomData.PullPriority.hasInDorm);
                }
            } catch (err) {
                console.error("Error fetching room data:", err);
            }
        };

        fetchRoomInfo();
    }, [selectedID, refreshKey, getRoomUUIDFromUserID, handleErrorFromTokenExpiry]);

    useEffect(() => {
        const storedCredentials = localStorage.getItem("jwt");
        if (storedCredentials) {
            setCredentials(storedCredentials);
        }
    }, [setCredentials]);

    const handleSuccess = (credentialResponse) => {
        if (!credentialResponse?.credential) return;
        // decode the credential
        const decoded = jwtDecode(credentialResponse.credential);
        setCredentials(credentialResponse.credential);
        localStorage.setItem("jwt", credentialResponse.credential);

        // Immediately query for user by email
        if (decoded.email) {
            fetchUserByEmail(decoded.email);
        }

        const hideWelcomeFAQ = localStorage.getItem("hideWelcomeFAQ");
        if (!hideWelcomeFAQ) {
            setShowFAQModal(true);
        }
    };

    // Function to fetch user by email
    const fetchUserByEmail = async (email) => {
        try {
            const response = await fetch(
                `${process.env.REACT_APP_API_URL}/users/email?email=${encodeURIComponent(email)}`,
                {
                    headers: {
                        Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                    },
                }
            );

            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }

            const data = await response.json();
            if (handleErrorFromTokenExpiry(data)) {
                return;
            }

            // If user found in DB, set userID immediately
            if (data.found) {
                setUserID(data.user.Id.toString());
                localStorage.setItem("userID", data.user.Id.toString());
                console.log("User found in database:", data.user.Id);

                // Also set as selectedID if there isn't one already
                if (!selectedID) {
                    setSelectedID(data.user.Id.toString());
                    localStorage.setItem("selectedID", data.user.Id.toString());
                }
            } else {
                console.log("User not found in database, continuing as guest");
                // User not in database, will be treated as guest
                setUserID(null);
                localStorage.removeItem("userID");
            }
        } catch (err) {
            console.error("Error fetching user by email:", err);
        }
    };

    // Store userID in localStorage when userMap is loaded - this is a fallback if fetching by email fails
    useEffect(() => {
        if (credentials && userMap) {
            const decodedToken = jwtDecode(credentials);
            console.log("Attempting to store userID:");
            console.log("Email from token:", decodedToken.email);
            console.log("UserMap available:", Object.keys(userMap).length);
            const userId = Object.keys(userMap || {}).find(
                (id) => userMap[id].Email === decodedToken.email || id === "701" // Temporary fix for test data
            );
            console.log("Found userId:", userId);
            if (userId) {
                localStorage.setItem("userID", userId);
                console.log("Stored userID in localStorage:", userId);
            }
        }
    }, [credentials, userMap]);

    const handleError = () => {
        // commented console.log ('Login Failed');
        // Optionally, handle login failure (e.g., by clearing stored credentials)
    };

    const fetchUserData = useCallback(
        async (userId) => {
            if (!userId || !localStorage.getItem("jwt")) return null;
            try {
                const response = await fetch(`${process.env.REACT_APP_API_URL}/users/${userId}`, {
                    headers: {
                        Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                    },
                });
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const data = await response.json();
                if (handleErrorFromTokenExpiry(data)) {
                    return null;
                }
                return data;
            } catch (err) {
                console.error("Error fetching user data:", err);
                return null;
            }
        },
        [handleErrorFromTokenExpiry]
    );

    useEffect(() => {
        const updateSelectedUserData = async () => {
            // Only proceed if selectedID exists and is in userMap
            if (selectedID && userMap && userMap[selectedID]) {
                const data = await fetchUserData(selectedID);
                if (data) {
                    setSelectedUserData(data);
                    if (data && data.RoomUUID && data.RoomUUID !== "00000000-0000-0000-0000-000000000000") {
                        try {
                            const roomResponse = await fetch(
                                `${process.env.REACT_APP_API_URL}/rooms/${data.RoomUUID}`,
                                {
                                    headers: {
                                        Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                                    },
                                }
                            );
                            if (!roomResponse.ok) {
                                throw new Error(`HTTP error! status: ${roomResponse.status}`);
                            }
                            const roomData = await roomResponse.json();
                            if (handleErrorFromTokenExpiry(roomData)) {
                                return;
                            }
                            const roomDisplay = `${roomData.DormName} ${roomData.RoomID}`;
                            setSelectedUserRoom(roomDisplay);
                        } catch (err) {
                            console.error("Error fetching room data:", err);
                            setSelectedUserRoom("no room yet");
                        }
                    } else {
                        setSelectedUserRoom("no room yet");
                    }
                }
            } else {
                // Reset data if no valid user is selected
                setSelectedUserData(null);
                setSelectedUserRoom("Unselected");
            }
        };
        updateSelectedUserData();
    }, [selectedID, refreshKey, userMap, fetchUserData, handleErrorFromTokenExpiry]);

    const getRoomObjectFromUserID = (userID) => {
        if (rooms) {
            for (let room of rooms) {
                if (room.Occupants && room.Occupants.includes(Number(userID))) {
                    return room;
                }
            }
        }
        return null;
    };

    const canUserToggleInDorm = (userID) => {
        // Safety check - if userID is invalid or not in userMap, return -1 (no toggle)
        if (!userID || !userMap || !userMap[userID]) {
            return -1;
        }

        userID = Number(userID);
        const usersRoom = getRoomObjectFromUserID(userID);
        if (!userMap) {
            return -1;
        }

        if (!usersRoom) {
            if (dormMapping[userMap[userID].InDorm]) {
                return 0;
            }
            return -1;
        }
        if (userMap[userID].InDorm === 0) {
            return -1;
        }
        if (usersRoom.MaxOccupancy > 1) {
            return 0;
        }

        if (dormMapping[userMap[userID].InDorm] === usersRoom.DormName) {
            return 1;
        } else if (dormMapping[userMap[userID].InDorm]) {
            return 0;
        }
        return -1;
    };

    // Save state to localStorage whenever it changes
    useEffect(() => {
        localStorage.setItem("activeTab", activeTab);
    }, [activeTab]);

    const handleNameChange = (selectedOption) => {
        if (selectedOption) {
            // Immediately unfocus search when an option is selected
            setIsSearchFocused(false);

            setLastSelectedID(selectedID); // Store the previous selection
            setSelectedID(selectedOption.value);
            // Immediately trigger an update of the user's room info when they are selected
            if (userMap && userMap[selectedOption.value]) {
                const updateRoomInfoImmediately = async () => {
                    const data = await fetchUserData(selectedOption.value);
                    if (data && data.RoomUUID && data.RoomUUID !== "00000000-0000-0000-0000-000000000000") {
                        try {
                            const roomResponse = await fetch(
                                `${process.env.REACT_APP_API_URL}/rooms/${data.RoomUUID}`,
                                {
                                    headers: {
                                        Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                                    },
                                }
                            );
                            if (roomResponse.ok) {
                                const roomData = await roomResponse.json();
                                if (!handleErrorFromTokenExpiry(roomData)) {
                                    const roomDisplay = `${roomData.DormName} ${roomData.RoomID}`;
                                    setSelectedUserRoom(roomDisplay);
                                }
                            }
                        } catch (err) {
                            console.error("Error fetching room data:", err);
                        }
                    } else {
                        setSelectedUserRoom("no room yet");
                    }
                };
                updateRoomInfoImmediately();
            }
        } else {
            setSelectedID(userID);
        }
    };

    const handleTabClick = (tab) => {
        setActiveTab(tab);
    };

    const handleForfeit = () => {
        // Skip if no selectedID
        if (!selectedID) return;

        // Get the room UUID for the selected user
        const roomUUID = getRoomUUIDFromUserID(selectedID);
        if (!roomUUID) {
            console.error("No room UUID found for user");
            return;
        }

        if (localStorage.getItem("jwt")) {
            fetch(`${process.env.REACT_APP_API_URL}/rooms/indorm/${roomUUID}`, {
                method: "POST",
                headers: {
                    Authorization: `Bearer ${localStorage.getItem("jwt")}`,
                },
            })
                .then((res) => {
                    return res.json();
                })
                .then((data) => {
                    setRefreshKey((prevKey) => prevKey + 1);
                    if (handleErrorFromTokenExpiry(data)) {
                        return;
                    }
                    setIsInDorm((prev) => !prev);
                })
                .catch((err) => {
                    console.error("Error forfeiting in-dorm:", err);
                });
        }
    };

    const handleTransitionStart = () => {
        setIsTransitioning(true);
    };

    const handleTransitionEnd = () => {
        setIsTransitioning(false);
    };

    // Check if current user is an admin
    const isAdmin = () => {
        if (!credentials) return false;
        try {
            const decoded = jwtDecode(credentials);
            return decoded.email === "tlam@g.hmc.edu" || decoded.email === "smao@g.hmc.edu";
        } catch (error) {
            return false;
        }
    };

    return (
        <div className={`main-content ${isTransitioning ? "transition-active" : ""}`}>
            {credentials && <Navbar />}

            <div className="content-wrapper">
                {!credentials && (
                    <section className="login-section">
                        <div className="login-card">
                            <img src="./digidraw.ico" alt="DigiDraw Logo" className="logo" />
                            <h1 className="title">Welcome to DigiDraw!</h1>
                            <p className="subtitle">Sign in with your school Google account to continue.</p>
                            <div className="google-login-wrapper">
                                <GoogleLogin
                                    auto_select={true}
                                    onSuccess={handleSuccess}
                                    onError={handleError}
                                    className="google-login"
                                />
                            </div>
                        </div>
                    </section>
                )}

                {credentials &&
                    notifications.map((notification, index) => (
                        <div
                            key={index}
                            className="notification is-info"
                            style={{ marginLeft: "20px", marginRight: "20px" }}
                        >
                            <button className="delete" onClick={() => handleCloseNotification(notification)}></button>
                            {notification}
                        </div>
                    ))}

                {credentials && (
                    <section
                        className="section search-section"
                        style={{
                            paddingTop: "1rem",
                            paddingBottom: "1rem",
                            borderBottom: "1px solid var(--border-color)",
                        }}
                    >
                        <div style={{ textAlign: "center" }}>
                            <h2 className="subtitle mb-4">Look up a student's room and draw number</h2>
                            <div className="search-container">
                                <div style={{ width: "var(--component-width)" }}>
                                    <Select
                                        className="react-select"
                                        classNamePrefix="react-select"
                                        placeholder="Search for a student..."
                                        onChange={handleNameChange}
                                        onFocus={() => {
                                            // Store the current selection before clearing it
                                            if (selectedID) {
                                                setLastSelectedID(selectedID);
                                            }
                                            setIsSearchFocused(true);
                                        }}
                                        onBlur={() => {
                                            setIsSearchFocused(false);
                                            if (!selectedID && lastSelectedID) {
                                                // Restore the last selection when no new selection is made
                                                setSelectedID(lastSelectedID);
                                            } else if (!selectedID && userID) {
                                                setSelectedID(userID);
                                            }
                                        }}
                                        menuPortalTarget={document.body}
                                        styles={{
                                            menuPortal: (base) => ({ ...base, zIndex: 9999 }),
                                            container: (base) => ({
                                                ...base,
                                                width: "100%",
                                                minWidth: "var(--min-component-width)",
                                            }),
                                            control: (base) => ({
                                                ...base,
                                                minWidth: "var(--min-component-width)",
                                                width: "100%",
                                            }),
                                        }}
                                        options={userOptions}
                                        value={
                                            isSearchFocused
                                                ? null // When focused, show empty input
                                                : selectedID && userMap && userMap[selectedID]
                                                  ? {
                                                        value: selectedID,
                                                        label: `${userMap[selectedID].FirstName} ${userMap[selectedID].LastName}`,
                                                    }
                                                  : null
                                        }
                                        openMenuOnFocus={true}
                                        isClearable={true}
                                    />
                                </div>
                                <div
                                    onClick={() =>
                                        selectedUserRoom !== `no room yet`
                                            ? handleTakeMeThere(selectedUserRoom, false)
                                            : null
                                    }
                                    className={`info-display ${selectedUserRoom !== `no room yet` ? "clickable" : "non-clickable"}`}
                                    title={
                                        selectedUserRoom !== `no room yet`
                                            ? `View room: ${selectedUserRoom}`
                                            : "No room assigned yet"
                                    }
                                >
                                    {selectedID && userMap && userMap[selectedID] ? (
                                        <>
                                            <span style={{ fontWeight: "500" }}>
                                                {userMap[selectedID].Preplaced
                                                    ? "Preplaced"
                                                    : `${userMap[selectedID].Year.charAt(0).toUpperCase() + userMap[selectedID].Year.slice(1)} #${userMap[selectedID].DrawNumber}`}
                                            </span>
                                            <span className="separator">•</span>
                                            <span style={{ color: "var(--text-color)" }}>
                                                {selectedUserRoom !== `no room yet` ? selectedUserRoom : "no room yet"}
                                            </span>
                                        </>
                                    ) : (
                                        <span className="has-text-grey">Student info will appear here</span>
                                    )}
                                </div>
                            </div>

                            {canUserToggleInDorm(selectedID) === 1 && (
                                <div
                                    style={{
                                        paddingTop: "1rem",
                                        display: "flex",
                                        alignItems: "center",
                                        justifyContent: "center",
                                        gap: "10px",
                                    }}
                                >
                                    <label htmlFor="toggleInDorm" className="checkbox-label">
                                        In-Dorm Forfeited
                                    </label>
                                    <input
                                        type="checkbox"
                                        className="switch"
                                        id="toggleInDorm"
                                        name="toggleInDorm"
                                        checked={isInDorm}
                                        onChange={handleForfeit}
                                    />
                                    <label htmlFor="toggleInDorm" className="checkbox-label">
                                        In-Dorm Kept
                                    </label>
                                </div>
                            )}
                            {canUserToggleInDorm(selectedID) === 0 && (
                                <div style={{ paddingTop: "1rem" }}>
                                    <p className="checkbox-label">
                                        To forfeit in-dorm, pull into a dorm for which you don't have in-dorm, or pull
                                        into {dormMapping[userMap[selectedID].InDorm]} and toggle the switch.
                                    </p>
                                </div>
                            )}
                        </div>
                    </section>
                )}

                {credentials && currPage === "Home" && (
                    <section className="section">
                        <div className="tabs is-centered">
                            <ul>
                                {gridData.length === 9 &&
                                    gridData
                                        .sort((a, b) => {
                                            const dormToNumber = {
                                                East: 1,
                                                North: 2,
                                                South: 3,
                                                West: 4,
                                                Atwood: 5,
                                                Sontag: 6,
                                                Case: 7,
                                                Drinkward: 8,
                                                Linde: 9,
                                            };
                                            return dormToNumber[a.dormName] - dormToNumber[b.dormName];
                                        })
                                        .map((dorm) => (
                                            <li
                                                key={dorm.dormName}
                                                className={activeTab === dorm.dormName ? "is-active" : ""}
                                                onClick={() => handleTabClick(dorm.dormName)}
                                            >
                                                <button type="button">{dorm.dormName}</button>
                                            </li>
                                        ))}
                            </ul>
                        </div>

                        <div className="columns is-centered" style={{ width: "100%", margin: 0 }}>
                            {!showFloorplans && (
                                <div className="column is-full">
                                    <div className="floor-content-wrapper">
                                        <TransitionGroup className="transition-group">
                                            <CSSTransition
                                                key={activeTab}
                                                timeout={300}
                                                classNames="crossfade"
                                                onEnter={handleTransitionStart}
                                                onExited={handleTransitionEnd}
                                            >
                                                <div
                                                    className="crossfade-wrapper"
                                                    style={{ backgroundColor: "var(--body-bg)" }}
                                                >
                                                    <div
                                                        className="columns is-centered is-multiline"
                                                        style={{ margin: 0, backgroundColor: "var(--body-bg)" }}
                                                    >
                                                        {gridData
                                                            .filter((dorm) => dorm.dormName === activeTab)
                                                            .flatMap((dorm) => dorm.floors)
                                                            .map((_, floorIndex) =>
                                                                activeTab === "Case" ? (
                                                                    floorIndex < 2 && (
                                                                        <div
                                                                            key={`${activeTab}-${floorIndex}`}
                                                                            className="column is-narrow"
                                                                            style={{
                                                                                padding: "0.75rem",
                                                                                backgroundColor: "var(--body-bg)",
                                                                            }}
                                                                        >
                                                                            <FloorDisplay
                                                                                gridData={gridData}
                                                                                filterCondition={(floorNumber) =>
                                                                                    floorNumber === floorIndex
                                                                                }
                                                                                showFloorplans={showFloorplans}
                                                                                activeTab={activeTab}
                                                                            />
                                                                        </div>
                                                                    )
                                                                ) : (
                                                                    <div
                                                                        key={`${activeTab}-${floorIndex}`}
                                                                        className="column is-narrow"
                                                                        style={{
                                                                            padding: "0.75rem",
                                                                            backgroundColor: "var(--body-bg)",
                                                                        }}
                                                                    >
                                                                        <FloorDisplay
                                                                            gridData={gridData}
                                                                            filterCondition={(floorNumber) =>
                                                                                floorNumber === floorIndex
                                                                            }
                                                                            showFloorplans={showFloorplans}
                                                                            activeTab={activeTab}
                                                                        />
                                                                    </div>
                                                                )
                                                            )}
                                                    </div>
                                                </div>
                                            </CSSTransition>
                                        </TransitionGroup>
                                    </div>
                                </div>
                            )}
                            {showFloorplans && (
                                <div className="column is-full">
                                    <div className="floor-content-wrapper">
                                        <TransitionGroup className="transition-group">
                                            <CSSTransition
                                                key={activeTab}
                                                timeout={300}
                                                classNames="crossfade"
                                                onEnter={handleTransitionStart}
                                                onExited={handleTransitionEnd}
                                            >
                                                <div
                                                    className="crossfade-wrapper"
                                                    style={{ backgroundColor: "var(--body-bg)" }}
                                                >
                                                    <div
                                                        className="floorplans-section"
                                                        style={{ backgroundColor: "var(--body-bg)" }}
                                                    >
                                                        {gridData
                                                            .filter((dorm) => dorm.dormName === activeTab)
                                                            .flatMap((dorm) => dorm.floors)
                                                            .map((_, floorIndex) => (
                                                                <div
                                                                    key={`${activeTab}-${floorIndex}`}
                                                                    className="floorplan-container"
                                                                >
                                                                    <FloorDisplay
                                                                        gridData={gridData}
                                                                        filterCondition={(floorNumber) =>
                                                                            floorNumber === floorIndex
                                                                        }
                                                                        showFloorplans={showFloorplans}
                                                                        activeTab={activeTab}
                                                                    />
                                                                    <img
                                                                        src={`./Floorplans/floorplans_${activeTab.toLowerCase()}_${floorIndex + 1}.png`}
                                                                        alt={`Floorplan for floor ${floorIndex}`}
                                                                        className="floorplan-image"
                                                                    />
                                                                </div>
                                                            ))}
                                                    </div>
                                                </div>
                                            </CSSTransition>
                                        </TransitionGroup>
                                    </div>
                                </div>
                            )}
                        </div>
                    </section>
                )}

                {currPage === "Recommendations" && (
                    <section className="section">
                        {/* <Recommendations gridData={gridData} setCurrPage={setCurrPage} /> */}
                    </section>
                )}

                {/* Search page */}
                {currPage === "Search" && credentials && <SearchPage />}

                {/* Admin dashboard page */}
                {currPage === "Admin" && isAdmin() && (
                    <section className="section">
                        <div className="container">
                            <h1 className="title has-text-centered">Admin Dashboard</h1>
                            <BlocklistManager />
                        </div>
                    </section>
                )}
            </div>

            {/* Fixed FAQ button */}
            {credentials && (
                <button
                    className="button is-primary is-rounded"
                    style={{
                        position: "fixed",
                        bottom: "20px",
                        right: "20px",
                        width: "40px",
                        height: "40px",
                        maxWidth: "40px",
                        padding: 0,
                        display: "flex",
                        alignItems: "center",
                        justifyContent: "center",
                        zIndex: 29,
                        boxShadow: "0 2px 8px rgba(0, 0, 0, 0.15)",
                    }}
                    onClick={() => setShowFAQModal(true)}
                    title="Show FAQ"
                >
                    <i className="fas fa-info"></i>
                </button>
            )}

            <Suspense fallback={null}>
                {isModalOpen && <BumpModal />}
                {isSuiteNoteModalOpen && <SuiteNoteModal />}
                {isFroshModalOpen && <BumpFroshModal />}
                {isSettingsModalOpen && <SettingsModal />}
                {isUserSettingsModalOpen && (
                    <UserSettingsModal isOpen={isUserSettingsModalOpen} onClose={() => setIsUserSettingsModalOpen(false)} />
                )}
                <FAQModal isOpen={showFAQModal} onClose={() => setShowFAQModal(false)} />
            </Suspense>
        </div>
    );
}

export default App;
