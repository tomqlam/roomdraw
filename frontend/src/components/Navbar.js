import { googleLogout } from '@react-oauth/google';
import { jwtDecode } from "jwt-decode";
import React, { useContext, useEffect, useState } from 'react';
import { MyContext } from '../MyContext';

function Navbar()
{
    const {
        userMap,
        credentials,
        setCredentials,
        lastRefreshedTime,
        setIsSettingsModalOpen,
        setIsUserSettingsModalOpen,
        handleErrorFromTokenExpiry,
        userID,
        handleTakeMeThere,
        refreshKey
    } = useContext(MyContext);

    const [isBurgerActive, setIsBurgerActive] = useState(false);
    const [currentUserData, setCurrentUserData] = useState(null);
    const [myRoom, setMyRoom] = useState("Unselected"); // to show what room current logged in user is in

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
        const updateCurrentUserData = async () =>
        {
            if (userID)
            {
                const data = await fetchUserData(userID);
                if (data)
                {
                    setCurrentUserData(data);
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
                            setMyRoom(roomDisplay);
                        } catch (err)
                        {
                            console.error('Error fetching room data:', err);
                            setMyRoom('no room yet');
                        }
                    } else
                    {
                        setMyRoom('no room yet');
                    }
                }
            }
        };
        updateCurrentUserData();
    }, [userID, refreshKey, handleErrorFromTokenExpiry]);

    return (
        <nav className="navbar" role="navigation" aria-label="main navigation">
            <div className="navbar-brand">
                <img src="./digidraw.ico" alt="DigiDraw Logo" style={{ height: '3rem' }} />
                <div className="navbar-last-refresh">
                    <h2>Last refresh: {lastRefreshedTime.toLocaleTimeString()}</h2>
                </div>
            </div>

            <a
                role="button"
                className={`navbar-burger ${isBurgerActive ? 'is-active' : ''}`}
                aria-label="menu"
                aria-expanded={isBurgerActive ? "true" : "false"}
                onClick={() => setIsBurgerActive(!isBurgerActive)}
            >
                <span aria-hidden="true"></span>
                <span aria-hidden="true"></span>
                <span aria-hidden="true"></span>
            </a>

            <div
                className={`navbar-menu-backdrop ${isBurgerActive ? 'is-active' : ''}`}
                onClick={() => setIsBurgerActive(false)}
            />

            <div className={`navbar-menu-items ${isBurgerActive ? 'is-active' : ''}`}>
                <div className="navbar-mobile-section">
                    {/* User Info Section */}
                    <div className="navbar-item user-info-wrapper">

                        <button
                            className="button is-light mobile-stack-item"
                            onClick={() => setIsUserSettingsModalOpen(true)}
                            title="User Settings"
                        >
                            <span className="icon">
                                <i className="fas fa-user-cog"></i>
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


                    </div>

                    <div className="navbar-item user-info-wrapper">
                        {userMap && (() =>
                        {
                            const decodedToken = jwtDecode(credentials);
                            const userId = Object.keys(userMap || {}).find(
                                id => userMap[id].Email === decodedToken.email
                            );
                            if (userId && userMap[userId])
                            {
                                return (
                                    <div className="info-display non-clickable mobile-stack-item" style={{ maxWidth: 'fit-content' }}>
                                        <span className="icon">
                                            <i className="fas fa-user"></i>
                                        </span>
                                        <span style={{ fontWeight: '500' }}>
                                            {userMap[userId].Year.charAt(0).toUpperCase() + userMap[userId].Year.slice(1)} #{userMap[userId].DrawNumber}
                                        </span>
                                    </div>
                                );
                            }
                            return null;
                        })()}
                        {userID && userMap && userMap[userID] && (
                            <div
                                onClick={() => myRoom !== `no room yet` ? handleTakeMeThere(myRoom, true) : null}
                                className={`info-display ${myRoom !== `no room yet` ? 'clickable' : 'non-clickable'} mobile-stack-item`}
                                title={myRoom !== `no room yet` ? `Go to my room: ${myRoom}` : 'No room assigned yet'}
                            >
                                <span className="icon">
                                    <i className="fas fa-home"></i>
                                </span>
                                <span style={{ fontWeight: '500' }}>
                                    {userMap[userID].Year.charAt(0).toUpperCase() + userMap[userID].Year.slice(1)} #{userMap[userID].DrawNumber}
                                </span>
                                <span className="separator">•</span>
                                <span style={{ color: 'var(--text-color)' }}>
                                    {myRoom !== `no room yet` ? myRoom : 'no room yet'}
                                </span>
                            </div>
                        )}
                    </div>

                    {/* Settings Section */}
                    <div className="navbar-item settings-wrapper">
                        <button
                            className="button is-primary mobile-stack-item"
                            onClick={() => setIsSettingsModalOpen(prev => !prev)}
                            title="Visual Settings"
                        >
                            <span className="icon">
                                <i className="fas fa-palette"></i>
                            </span>
                            <span>Visual Settings</span>
                        </button>
                    </div>

                    {/* Logout Section */}
                    <div className="navbar-item logout-wrapper">
                        <a className="button is-danger mobile-stack-item" onClick={handleLogout} title="Log Out">
                            <span className="icon">
                                <i className="fas fa-sign-out-alt"></i>
                            </span>
                            <span>Log Out</span>
                        </a>
                    </div>
                </div>
            </div>
        </nav>
    );
}

export default Navbar; 