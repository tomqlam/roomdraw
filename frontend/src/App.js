import '@fortawesome/fontawesome-free/css/all.min.css';
import { GoogleLogin, googleLogout } from '@react-oauth/google';
import 'bulma/css/bulma.min.css';
import { jwtDecode } from "jwt-decode";
import React, { useContext, useEffect, useState } from 'react';
import Select from 'react-select';
import BumpFroshModal from './BumpFroshModal';
import BumpModal from './BumpModal';
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
    } = useContext(MyContext);

    const [notifications, setNotifications] = useState([]);
    const [isUserSettingsModalOpen, setIsUserSettingsModalOpen] = useState(false);

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

    const [selectedUserRoom, setSelectedUserRoom] = useState("You are not in a room yet."); // to show what room current selected user is in
    const [myRoom, setMyRoom] = useState("You are not in a room yet."); // to show what room current logged in user is in
    const [isBurgerClicked, setIsBurgerClicked] = useState(false);
    const [isInDorm, setIsInDorm] = useState(true);

    useEffect(() =>
    {

        const thisRoom = getRoomObjectFromUserID(selectedID);
        if (thisRoom)
        {
            // commented console.log (thisRoom);
            // commented console.log (thisRoom.PullPriority);
            var pullPriority = thisRoom.PullPriority;
            if (pullPriority.inherited.valid)
            {
                pullPriority = pullPriority.inherited;
            }
            if (pullPriority.hasInDorm === true)
            {
                setIsInDorm(false);
            } else
            {
                setIsInDorm(true);
            }

        }

    }, [selectedID, rooms]);

    useEffect(() =>
    {
        const storedCredentials = localStorage.getItem('jwt');
        if (storedCredentials)
        {
            // commented console.log ("use effect");
            // commented console.log (storedCredentials);
            // commented console.log ("end use efect");
            setCredentials(storedCredentials);
        }
    }, []);

    // useEffect(() => {
    //   // Check for stored credentials on component mount
    //   const storedCredentials = localStorage.getItem('jwt');
    //   if (storedCredentials) {
    //     // If credentials exist, decode and set them
    //     setCredentials(storedCredentials);
    //   }
    // }, []);

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

    const getRoomObjectFromUserID = (userID) =>
    {
        if (rooms)
        {
            for (let room of rooms)
            {

                if (room.Occupants && room.Occupants.includes(Number(selectedID)))
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
        if (dormMapping[userMap[userID].InDorm] === usersRoom.DormName)
        {
            return 1;
        } else if (dormMapping[userMap[userID].InDorm])
        {
            return 0;
        }
        return -1;

    }



    useEffect(() =>
    {
        // updates room that the selected user is in every time the selected user or the room data changes
        if (!rooms || !Array.isArray(rooms))
        {
            return;
        }
        if (rooms)
        {
            for (let room of rooms)
            {

                if (room.Occupants && room.Occupants.includes(Number(selectedID)))
                {


                    setSelectedUserRoom(`${room.DormName} ${room.RoomID}`);
                    return;
                }
            }
            setSelectedUserRoom(`no room yet`);


        }
    }, [selectedID, rooms]);

    useEffect(() =>
    {
        // updates room that the logged in user is in every time the selected user or the room data changes
        if (!rooms || !Array.isArray(rooms))
        {
            return;
        }
        if (rooms)
        {
            for (let room of rooms)
            {

                if (room.Occupants && room.Occupants.includes(Number(userID)))
                {


                    setMyRoom(`${room.DormName} ${room.RoomID}`);
                    return;
                }
            }
            setMyRoom(`no room yet`);


        }
    }, [userID, rooms]);

    // Save state to localStorage whenever it changes
    useEffect(() =>
    {
        localStorage.setItem('activeTab', activeTab);
    }, [activeTab]);


    const handleNameChange = (newID) =>
    {
        setSelectedID(newID);
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
        const filteredFloors = gridData.flatMap(dorm => dorm.floors.filter(floor => filterCondition(floor.floorNumber)));

        return (
            <div className="column">
                <div style={showFloorplans ? { width: '50vw' } : {}}>
                    {gridData.map((dorm) => (
                        <div key={dorm.dormName} className={activeTab === dorm.dormName ? '' : 'is-hidden'}>
                            {dorm.floors
                                .filter((floor) => filterCondition(floor.floorNumber))
                                .sort((a, b) => Number(a.floorNumber) - Number(b.floorNumber))  // Convert to numbers before comparing
                                .map((floor, floorIndex) => (
                                    <div style={{ paddingBottom: 20 }} className="container" key={floorIndex}>
                                        <h2 className="subtitle has-text-centered">Floor {floor.floorNumber + 1}</h2>
                                        {floor.floorName && <p className="subtitle has-text-centered">{floor.floorName}</p>}
                                        <FloorGrid gridData={floor} />
                                    </div>
                                ))}
                        </div>
                    ))}
                </div>
            </div>
        );
    };
    const handleTakeMeThere = (myLocationString) =>
    {
        const words = myLocationString.split(' ');
        if (words.length === 2)
        {
            setActiveTab(words[0]);
        }

        // Assume `selectedID` is the ID of the selected room
        const roomUUID = getRoomUUIDFromUserID(selectedID); // Replace this with the actual function to get the room UUID

        // Delay the scrolling until after the tab has finished switching
        setTimeout(() =>
        {
            const roomRef = roomRefs.current[roomUUID];
            if (roomRef)
            {
                roomRef.scrollIntoView({ behavior: 'smooth' });
            }
        }, 0);
    }

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
                    // commented console.log ("BRUHMOMENT");
                    // commented console.log (thisRoom);
                    // commented console.log (thisRoom.PullPriority.hasInDorm);
                    setIsInDorm(prev => !prev);

                })
                .catch(err =>
                {
                    // commented console.log (err);
                })


        }
    }

    return (
        <div>
            <nav class="navbar" role="navigation" aria-label="main navigation">
                <div class="navbar-brand">
                    <a class="navbar-item" href="#"><img src="https://i.ibb.co/SyRVPQN/Screenshot-2023-12-26-at-10-14-31-PM.png" alt="Screenshot-2023-12-26-at-10-14-31-PM" border="0" /></a>

                    {/* <a role="button" class="navbar-burger" aria-label="menu" aria-expanded="false" data-target="navbarBasicExample" onClick={() => setIsBurgerClicked(true)}>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
            <span aria-hidden="true"></span>
          </a> */}
                    {(!credentials && window.innerWidth < 1024) &&
                        <GoogleLogin auto_select={true}
                            onSuccess={handleSuccess}
                            onError={handleError}
                        />}
                    {(credentials && window.innerWidth < 1024) && <a class="button is-danger" onClick={handleLogout}>
                        <strong>Log Out</strong>
                    </a>}
                </div>



                <div id="navbarBasicExample" class="navbar-menu">
                    <div class="navbar-start">
                        <div class="navbar-item">
                            <h2>Last refresh: {lastRefreshedTime.toLocaleTimeString()}</h2>
                        </div>
                    </div>

                    <div className="navbar-end">
                        <div className="navbar-item">
                            <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
                                {credentials && userMap && (() =>
                                {
                                    const decodedToken = jwtDecode(credentials);
                                    const userId = Object.keys(userMap || {}).find(
                                        id => userMap[id].Email === decodedToken.email
                                    );
                                    if (userId && userMap[userId])
                                    {
                                        return (
                                            <div className="info-display non-clickable" style={{ maxWidth: 'fit-content' }}>
                                                <span style={{ fontWeight: '500' }}>
                                                    {userMap[userId].Year.charAt(0).toUpperCase() + userMap[userId].Year.slice(1)} #{userMap[userId].DrawNumber}
                                                </span>
                                            </div>
                                        );
                                    }
                                    return null;
                                })()}

                                <button
                                    className="button is-light"
                                    onClick={() => setIsUserSettingsModalOpen(true)}
                                >
                                    <span className="icon">
                                        <i className="fas fa-user"></i>
                                    </span>
                                    <span>Welcome, {(() =>
                                    {
                                        if (!credentials) return '';
                                        const decodedToken = jwtDecode(credentials);
                                        const userId = Object.keys(userMap || {}).find(
                                            id => userMap[id].Email === decodedToken.email
                                        );
                                        return userId && userMap[userId] ? userMap[userId].FirstName : decodedToken.given_name;
                                    })()}</span>
                                </button>

                                <div
                                    onClick={() => myRoom !== `no room yet` ? handleTakeMeThere(myRoom) : null}
                                    className={`info-display ${myRoom !== `no room yet` ? 'clickable' : 'non-clickable'}`}
                                >
                                    {userID && userMap && userMap[userID] ? (
                                        <>
                                            <span style={{ fontWeight: '500' }}>
                                                {userMap[userID].Year.charAt(0).toUpperCase() + userMap[userID].Year.slice(1)} #{userMap[userID].DrawNumber}
                                            </span>
                                            <span className="separator">•</span>
                                            <span style={{ color: 'var(--text-color)' }}>
                                                {myRoom !== `no room yet` ? myRoom : 'no room yet'}
                                            </span>
                                        </>
                                    ) : (
                                        <span className="has-text-grey">Student info will appear here</span>
                                    )}
                                </div>

                                <button
                                    className="button is-primary"
                                    onClick={() => setIsSettingsModalOpen(prev => !prev)}
                                >
                                    <span className="icon">
                                        <i className="fas fa-palette"></i>
                                    </span>
                                    <span>Visual Settings</span>
                                </button>

                                <a className="button is-danger" onClick={handleLogout}>
                                    <span className="icon">
                                        <i className="fas fa-sign-out-alt"></i>
                                    </span>
                                    <span>Log Out</span>
                                </a>
                            </div>
                        </div>
                    </div>
                </div>
            </nav>
            {isModalOpen && <BumpModal />}
            {isSuiteNoteModalOpen && <SuiteNoteModal />}
            {isFroshModalOpen && <BumpFroshModal />}
            {isSettingsModalOpen && <SettingsModal />}
            {isUserSettingsModalOpen && <UserSettingsModal isOpen={isUserSettingsModalOpen} onClose={() => setIsUserSettingsModalOpen(false)} />}

            {/* {(credentials && showNotification) && <div class="notification is-link" style={{ marginLeft: '20px', marginRight: '20px' }}>
        <button class="delete" onClick={handleCloseNotification}></button>
        UI Update: 1) Preplaced rooms now appear slightly darker by default. 2) Unbumpable rooms are now always black.
      </div>} */}
            {credentials && notifications.map((notification, index) => (
                <div key={index} className="notification is-info" style={{ marginLeft: '20px', marginRight: '20px' }}>
                    <button className="delete" onClick={() => handleCloseNotification(notification)}></button>
                    {notification}
                </div>
            ))}




            {!credentials && <section class="section">
                <div style={{ textAlign: 'center' }}>
                    <h1 className="title">Welcome to DigiDraw!</h1>
                    <h2 className="subtitle">Please log in with your HMC email to continue.</h2>
                </div>
            </section>}
            {credentials && <section className="section" style={{ paddingTop: '1rem', paddingBottom: '1rem', borderBottom: '1px solid #eee', backgroundColor: '#f8f9fa' }}>
                <div style={{ textAlign: 'center' }}>
                    <h2 className="subtitle mb-4">Look up a student's room and draw number</h2>
                    <div className="search-container">
                        <div style={{ width: 'var(--component-width)' }}>
                            <Select
                                placeholder="Type a name to search..."
                                value={userMap && selectedID && userMap[selectedID] ? {
                                    value: selectedID,
                                    label: `${userMap[selectedID].FirstName} ${userMap[selectedID].LastName}`
                                } : null}
                                menuPortalTarget={document.body}
                                classNamePrefix="react-select"
                                styles={{
                                    menuPortal: base => ({ ...base, zIndex: 9999 })
                                }}
                                onChange={(selectedOption) => handleNameChange(selectedOption.value)}
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
                            />
                        </div>
                        <div
                            onClick={() => selectedUserRoom !== `no room yet` ? handleTakeMeThere(selectedUserRoom) : null}
                            className={`info-display ${selectedUserRoom !== `no room yet` ? 'clickable' : 'non-clickable'}`}
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

                    {canUserToggleInDorm(selectedID) !== -1 &&
                        <div>
                            <input
                                type="checkbox"
                                id="toggleInDorm"
                                name="toggleInDorm"
                                disabled={canUserToggleInDorm(selectedID) === 0}
                                checked={isInDorm}
                                onChange={handleForfeit}
                            />
                            {canUserToggleInDorm(selectedID) === 1 &&
                                <label htmlFor="toggleInDorm" style={{ marginLeft: '5px' }}>Forfeit In-Dorm for their current single</label>
                            }
                            {canUserToggleInDorm(selectedID) === 0 &&
                                <label htmlFor="toggleInDorm" style={{ marginLeft: '5px' }}>To forfeit their in-dorm, pull into a single.</label>
                            }
                        </div>
                    }
                </div>
            </section>}

            {(credentials && currPage === "Home") && <section class="section">


                <div className="tabs is-centered">
                    <ul>

                        {gridData.length === 9 && gridData.map((dorm) => (
                            (
                                <li
                                    key={dorm.dormName}
                                    className={activeTab === dorm.dormName ? 'is-active' : ''}
                                    onClick={() => handleTabClick(dorm.dormName)}
                                >
                                    <a>{dorm.dormName}</a>
                                </li>
                            )
                        ))}
                    </ul>
                </div>

                {userMap && <div class="columns">
                    {!showFloorplans && gridData
                        .filter(dorm => dorm.dormName === activeTab)
                        .flatMap(dorm => dorm.floors)
                        .map((_, floorIndex) => (
                            activeTab === "Case" ? (
                                floorIndex < 2 && <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                            ) : (
                                <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                            )
                        ))}
                    {showFloorplans && (
                        <div style={{ display: 'flex', flexDirection: 'column' }}>
                            {gridData
                                .filter(dorm => dorm.dormName === activeTab)
                                .flatMap(dorm => dorm.floors)
                                .map((_, floorIndex) => (
                                    <div style={{ display: 'flex', alignItems: 'center' }}>
                                        <FloorDisplay gridData={gridData} filterCondition={(floorNumber) => floorNumber === floorIndex} />
                                        <img
                                            src={`/Floorplans/floorplans_${activeTab.toLowerCase()}_${floorIndex + 1}.png`}
                                            alt={`Floorplan for floor ${floorIndex}`}
                                            style={{ maxWidth: '40vw' }} // Add this line
                                        />
                                    </div>
                                ))}
                        </div>
                    )}
                </div>}


            </section>}

            {currPage === "Recommendations" && <section class="section">
                {/* <Recommendations gridData={gridData} setCurrPage={setCurrPage} /> */}
            </section>}


            <footer class="footer">
                <div class="content has-text-centered">
                    <p>
                        <strong>Digital Draw</strong> by Serena Mao & Tom Lam. Email smao@g.hmc.edu or tlam@g.hmc.edu with questions.
                    </p>
                </div>
            </footer>


        </div>

    );
}


export default App;