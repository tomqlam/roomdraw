import '@fortawesome/fontawesome-free/css/all.min.css';
import { GoogleLogin, googleLogout } from '@react-oauth/google';
import 'bulma/css/bulma.min.css';
import { jwtDecode } from "jwt-decode";
import React, { useContext, useEffect, useState } from 'react';
import Select from 'react-select';
import { CSSTransition, TransitionGroup } from 'react-transition-group';
import BumpFroshModal from './BumpFroshModal';
import BumpModal from './BumpModal';
import Navbar from './components/Navbar';
import FloorGrid from './FloorGrid';
import { MyContext } from './MyContext';
import SettingsModal from './SettingsModal';
import './styles.css';
import SuiteNoteModal from './SuiteNoteModal';
import UserSettingsModal from './UserSettingsModal';


function App()
{

    const {
        currPage,
        setCurrPage,
        gridData,
        userMap,
        isModalOpen,
        dormMapping,
        onlyShowBumpableRooms,
        setOnlyShowBumpableRooms,
        getNameById,
        selectedID,
        setSelectedID,
        rooms,
        isSuiteNoteModalOpen,
        credentials,
        setCredentials,
        lastRefreshedTime,
        activeTab,
        setActiveTab,
        isFroshModalOpen,
        getRoomUUIDFromUserID,
        roomRefs,
        setRefreshKey,
        handleErrorFromTokenExpiry,
        isSettingsModalOpen,
        setIsSettingsModalOpen,
        showFloorplans,
        setShowFloorplans,
        userID,
        setUserID,
        refreshKey,
        isDarkMode,
        isUserSettingsModalOpen,
        setIsUserSettingsModalOpen,
        handleTakeMeThere
    } = useContext(MyContext);

    const [notifications, setNotifications] = useState([]);
    const [isTransitioning, setIsTransitioning] = useState(false);
    const [selectedUserData, setSelectedUserData] = useState(null);
    const [selectedUserRoom, setSelectedUserRoom] = useState("Unselected"); // to show what room current selected user is in
    const [isBurgerActive, setIsBurgerActive] = useState(false);
    const [isInDorm, setIsInDorm] = useState(true);

    useEffect(() =>
    {
        const closedNotifications = JSON.parse(localStorage.getItem('closedNotifications')) || [];
        const newNotifications = [
            'Reminder: to show/hide floorplans, click Settings in the top right corner and toggle the checkbox.',
            // 'Notification 2', 
        ].filter(notification => !closedNotifications.includes(notification));
        setNotifications(newNotifications);
    }, []);

    const handleCloseNotification = (notification) =>
    {
        // Store the notification state in local storage when the user closes the notification
        const closedNotifications = JSON.parse(localStorage.getItem('closedNotifications')) || [];
        closedNotifications.push(notification);
        localStorage.setItem('closedNotifications', JSON.stringify(closedNotifications));
        setNotifications(notifications.filter(n => n !== notification));
    };

    useEffect(() =>
    {
        const fetchRoomInfo = async () =>
        {
            const roomUUID = getRoomUUIDFromUserID(selectedID);
            if (!roomUUID) return;

            try
            {
                const response = await fetch(`/rooms/${roomUUID}`, {
                    headers: {
                        'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                    }
                });
                if (!response.ok)
                {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                const roomData = await response.json();
                if (handleErrorFromTokenExpiry(roomData))
                {
                    return;
                }

                // Check if the room has inherited priority
                if (roomData.pull_priority.inherited.valid)
                {
                    setIsInDorm(roomData.pull_priority.inherited.hasInDorm);
                } else
                {
                    setIsInDorm(roomData.pull_priority.hasInDorm);
                }
            } catch (err)
            {
                console.error('Error fetching room data:', err);
            }
        };

        fetchRoomInfo();
    }, [selectedID, refreshKey]);

    useEffect(() =>
    {
        const storedCredentials = localStorage.getItem('jwt');
        if (storedCredentials)
        {
            setCredentials(storedCredentials);
        }
    }, []);

    const handleSuccess = (credentialResponse) =>
    {
        if (!credentialResponse?.credential) return;
        // decode the credential
        const decoded = jwtDecode(credentialResponse.credential);
        setCredentials(credentialResponse.credential);
        localStorage.setItem('jwt', credentialResponse.credential);
    };

    // Store userID in localStorage when userMap is loaded
    useEffect(() =>
    {
        if (credentials && userMap)
        {
            const decodedToken = jwtDecode(credentials);
            console.log('Attempting to store userID:');
            console.log('Email from token:', decodedToken.email);
            console.log('UserMap available:', Object.keys(userMap).length);
            const userId = Object.keys(userMap || {}).find(
                id => userMap[id].Email === decodedToken.email || id === '701'  // Temporary fix for test data
            );
            console.log('Found userId:', userId);
            if (userId)
            {
                localStorage.setItem('userID', userId);
                console.log('Stored userID in localStorage:', userId);
            }
        }
    }, [credentials, userMap]);

    const handleError = () =>
    {
        // commented console.log ('Login Failed');
        // Optionally, handle login failure (e.g., by clearing stored credentials)
    };

    const handleLogout = () =>
    {
        setCredentials(null);
        localStorage.removeItem('jwt');
        localStorage.removeItem('userID');
        googleLogout();
    };

    const fetchUserData = async (userId) =>
    {
        if (!userId || !localStorage.getItem('jwt')) return null;
        try
        {
            const response = await fetch(`/users/${userId}`, {
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                }
            });
            if (!response.ok)
            {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const data = await response.json();
            if (handleErrorFromTokenExpiry(data))
            {
                return null;
            }
            return data;
        } catch (err)
        {
            console.error('Error fetching user data:', err);
            return null;
        }
    };

    useEffect(() =>
    {
        const updateSelectedUserData = async () =>
        {
            if (selectedID)
            {
                const data = await fetchUserData(selectedID);
                if (data)
                {
                    setSelectedUserData(data);
                    if (data.RoomUUID)
                    {
                        try
                        {
                            const roomResponse = await fetch(`/rooms/${data.RoomUUID}`, {
                                headers: {
                                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                                }
                            });
                            if (!roomResponse.ok)
                            {
                                throw new Error(`HTTP error! status: ${roomResponse.status}`);
                            }
                            const roomData = await roomResponse.json();
                            if (handleErrorFromTokenExpiry(roomData))
                            {
                                return;
                            }
                            const roomDisplay = `${roomData.DormName} ${roomData.RoomID}`;
                            setSelectedUserRoom(roomDisplay);
                        } catch (err)
                        {
                            console.error('Error fetching room data:', err);
                            setSelectedUserRoom('no room yet');
                        }
                    } else
                    {
                        setSelectedUserRoom('no room yet');
                    }
                }
            }
        };
        updateSelectedUserData();
    }, [selectedID, refreshKey]);

    const getRoomObjectFromUserID = (userID) =>
    {
        if (rooms)
        {
            for (let room of rooms)
            {

                if (room.Occupants && room.Occupants.includes(Number(userID)))
                {
                    return room;
                }
            }
        }
        return null;
    }

    const canUserToggleInDorm = (userID) =>
    {
        userID = Number(userID);
        const usersRoom = getRoomObjectFromUserID(userID);
        if (!userMap)
        {
            return -1;
        }

        if (!usersRoom)
        {
            if (dormMapping[userMap[userID].InDorm])
            {
                return 0;
            }
            return -1;
        }
        if (userMap[userID].InDorm === 0)
        {
            return -1;
        }
        if (usersRoom.MaxOccupancy > 1)
        {
            return 0;
        }

        if (dormMapping[userMap[userID].InDorm] === usersRoom.DormName)
        {
            return 1;
        } else if (dormMapping[userMap[userID].InDorm])
        {
            return 0;
        }
        return -1;

    }

    // Save state to localStorage whenever it changes
    useEffect(() =>
    {
        localStorage.setItem('activeTab', activeTab);
    }, [activeTab]);


    const handleNameChange = (selectedOption) =>
    {
        if (selectedOption)
        {
            setSelectedID(selectedOption.value);
        } else
        {
            setSelectedID(userID);
        }
    };

    const handleTabClick = (tab) =>
    {
        setActiveTab(tab);
    };

    function getDrawNumberAndYear(id)
    {
        // Find the drawNumber in laymans terms with the given id, including in-dorm status
        // ex: given 2, returns Sophomore 46
        if (!userMap)
        {
            return "Loading...";
        } else if (userMap[id].InDorm && userMap[id].InDorm !== 0)
        {
            // has in dorm
            return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber} with ${dormMapping[userMap[id].InDorm]} In-Dorm`;
        }
        return `${userMap[id].Year.charAt(0).toUpperCase() + userMap[id].Year.slice(1)} ${userMap[id].DrawNumber}`
    }



    // Component for each floor, to show even and odd floors separately
    const FloorDisplay = ({ gridData, filterCondition }) =>
    {
        return (
            <div className="column">
                <div style={showFloorplans ? { width: '50vw' } : {}}>
                    {gridData.map((dorm) => (
                        <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
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

    const handleForfeit = () =>
    {
        if (localStorage.getItem('jwt'))
        {
            fetch(`/rooms/indorm/${getRoomUUIDFromUserID(selectedID)}`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
                },
            })
                .then(res =>
                {
                    return res.json();
                })
                .then(data =>
                {
                    setRefreshKey(prevKey => prevKey + 1);
                    if (handleErrorFromTokenExpiry(data))
                    {
                        return;
                    };
                    const thisRoom = getRoomObjectFromUserID(selectedID);
                    setIsInDorm(prev => !prev);

                })
                .catch(err =>
                {
                    console.error('Error forfeiting in-dorm:', err);
                })


        }
    }

    const handleTransitionStart = () =>
    {
        setIsTransitioning(true);
    };

    const handleTransitionEnd = () =>
    {
        setIsTransitioning(false);
    };

    return (
        <div className={`main-content ${isTransitioning ? 'transition-active' : ''}`}>
            {credentials && <Navbar />}

            <div className="content-wrapper">
                {!credentials && (
                    <section className="login-section">
                        <div className="login-card">
                            <img src="./digidraw.ico" alt="DigiDraw Logo" className="logo" />
                            <h1 className="title">Welcome to DigiDraw!</h1>
                            <p className="subtitle">
                                Sign in with your school Google account to continue.
                            </p>
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

                {credentials && notifications.map((notification, index) => (
                    <div key={index} className="notification is-info" style={{ marginLeft: '20px', marginRight: '20px' }}>
                        <button className="delete" onClick={() => handleCloseNotification(notification)}></button>
                        {notification}
                    </div>
                ))}

                {credentials && <section className="section search-section" style={{ paddingTop: '1rem', paddingBottom: '1rem', borderBottom: '1px solid var(--border-color)' }}>
                    <div style={{ textAlign: 'center' }}>
                        <h2 className="subtitle mb-4">Look up a student's room and draw number</h2>
                        <div className="search-container">
                            <div style={{ width: 'var(--component-width)' }}>
                                <Select
                                    className="react-select"
                                    classNamePrefix="react-select"
                                    placeholder="Student name..."
                                    onChange={handleNameChange}
                                    menuPortalTarget={document.body}
                                    styles={{
                                        menuPortal: base => ({ ...base, zIndex: 9999 })
                                    }}
                                    options={userMap ? Object.keys(userMap)
                                        .sort((a, b) =>
                                        {
                                            const nameA = `${userMap[a].FirstName} ${userMap[a].LastName}`;
                                            const nameB = `${userMap[b].FirstName} ${userMap[b].LastName}`;
                                            return nameA.localeCompare(nameB);
                                        })
                                        .filter((key) => Number(userMap[key].Year) !== 0)
                                        .map((key) => ({
                                            value: key,
                                            label: `${userMap[key].FirstName} ${userMap[key].LastName}`
                                        })) : []}
                                    value={selectedID && userMap && userMap[selectedID] ?
                                        { value: selectedID, label: `${userMap[selectedID].FirstName} ${userMap[selectedID].LastName}` } : null}
                                />
                            </div>
                            <div
                                onClick={() => selectedUserRoom !== `no room yet` ? handleTakeMeThere(selectedUserRoom, false) : null}
                                className={`info-display ${selectedUserRoom !== `no room yet` ? 'clickable' : 'non-clickable'}`}
                                title={selectedUserRoom !== `no room yet` ? `View room: ${selectedUserRoom}` : 'No room assigned yet'}
                            >
                                {selectedID && userMap && userMap[selectedID] ? (
                                    <>
                                        <span style={{ fontWeight: '500' }}>
                                            {userMap[selectedID].Year.charAt(0).toUpperCase() + userMap[selectedID].Year.slice(1)} #{userMap[selectedID].DrawNumber}
                                        </span>
                                        <span className="separator">•</span>
                                        <span style={{ color: 'var(--text-color)' }}>
                                            {selectedUserRoom !== `no room yet` ? selectedUserRoom : 'no room yet'}
                                        </span>
                                    </>
                                ) : (
                                    <span className="has-text-grey">Student info will appear here</span>
                                )}
                            </div>
                        </div>

                        {canUserToggleInDorm(selectedID) === 1 &&
                            <div style={{ paddingTop: '1rem', display: 'flex', alignItems: 'center', justifyContent: 'center', gap: '10px' }}>
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
                        }
                        {canUserToggleInDorm(selectedID) === 0 &&
                            <div style={{ paddingTop: '1rem' }}>
                                <p className="checkbox-label">
                                    To forfeit in-dorm, pull into a single in {dormMapping[userMap[selectedID].InDorm]}
                                </p>
                            </div>
                        }
                    </div>
                </section>}

                {(credentials && currPage === "Home") && <section className="section">
                    <div className="tabs is-centered">
                        <ul>

                            {gridData.length === 9 && gridData
                                .sort((a, b) =>
                                {
                                    const dormToNumber = {
                                        "East": 1,
                                        "North": 2,
                                        "South": 3,
                                        "West": 4,
                                        "Atwood": 5,
                                        "Sontag": 6,
                                        "Case": 7,
                                        "Drinkward": 8,
                                        "Linde": 9
                                    };
                                    return dormToNumber[a.dormName] - dormToNumber[b.dormName];
                                })
                                .map((dorm) => (
                                    <li
                                        key={dorm.dormName}
                                        className={activeTab === dorm.dormName ? 'is-active' : ''}
                                        onClick={() => handleTabClick(dorm.dormName)}
                                    >
                                        <a>{dorm.dormName}</a>
                                    </li>
                                ))}
                        </ul>
                    </div>

                    {userMap && <div className="columns is-centered" style={{ width: '100%', margin: 0 }}>
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
                                            <div className="crossfade-wrapper" style={{ backgroundColor: 'var(--body-bg)' }}>
                                                <div className="columns is-centered is-multiline" style={{ margin: 0, backgroundColor: 'var(--body-bg)' }}>
                                                    {gridData
                                                        .filter(dorm => dorm.dormName === activeTab)
                                                        .flatMap(dorm => dorm.floors)
                                                        .map((_, floorIndex) => (
                                                            activeTab === "Case" ? (
                                                                floorIndex < 2 && (
                                                                    <div key={`${activeTab}-${floorIndex}`} className="column is-narrow" style={{ padding: '0.75rem', backgroundColor: 'var(--body-bg)' }}>
                                                                        <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                                                                    </div>
                                                                )
                                                            ) : (
                                                                <div key={`${activeTab}-${floorIndex}`} className="column is-narrow" style={{ padding: '0.75rem', backgroundColor: 'var(--body-bg)' }}>
                                                                    <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                                                                </div>
                                                            )
                                                        ))}
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
                                            <div className="crossfade-wrapper" style={{ backgroundColor: 'var(--body-bg)' }}>
                                                <div className="floorplans-section" style={{ backgroundColor: 'var(--body-bg)' }}>
                                                    {gridData
                                                        .filter(dorm => dorm.dormName === activeTab)
                                                        .flatMap(dorm => dorm.floors)
                                                        .map((_, floorIndex) => (
                                                            <div key={`${activeTab}-${floorIndex}`} className="floorplan-container">
                                                                <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                                                                <img
                                                                    src={`/Floorplans/floorplans_${activeTab.toLowerCase()}_${floorIndex + 1}.png`}
                                                                    alt={`Floorplan for floor ${floorIndex}`}
                                                                    className="floorplan-image"
                                                                    style={{ maxWidth: '40vw' }}
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
                    </div>}
                </section>}

                {currPage === "Recommendations" && <section className="section">
                    {/* <Recommendations gridData={gridData} setCurrPage={setCurrPage} /> */}
                </section>}
            </div>

            {isModalOpen && <BumpModal />}
            {isSuiteNoteModalOpen && <SuiteNoteModal />}
            {isFroshModalOpen && <BumpFroshModal />}
            {isSettingsModalOpen && <SettingsModal />}
            {isUserSettingsModalOpen && <UserSettingsModal isOpen={isUserSettingsModalOpen} onClose={() => setIsUserSettingsModalOpen(false)} />}
        </div>
    );
}


export default App;