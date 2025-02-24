import { jwtDecode } from "jwt-decode";
import React, { useContext, useEffect, useState } from 'react';
import { MyContext } from './MyContext';

const UserSettingsModal = ({ isOpen, onClose }) =>
{
    const { credentials } = useContext(MyContext);
    const [notificationsEnabled, setNotificationsEnabled] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() =>
    {
        if (credentials && isOpen)
        {
            const decodedToken = jwtDecode(credentials);
            const userEmail = decodedToken.email;

            // Fetch current notification preferences
            fetch(`/users/notifications?email=${encodeURIComponent(userEmail)}`, {
                headers: {
                    'Authorization': `Bearer ${credentials}`,
                }
            })
                .then(res => res.json())
                .then(data =>
                {
                    setNotificationsEnabled(data.enabled);
                    setLoading(false);
                })
                .catch(err =>
                {
                    console.error('Failed to fetch notification preferences:', err);
                    setLoading(false);
                });
        }
    }, [credentials, isOpen]);

    const handleToggleNotifications = () =>
    {
        if (!credentials) return;

        const decodedToken = jwtDecode(credentials);
        const userEmail = decodedToken.email;

        setLoading(true);
        fetch(`/users/notifications`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${credentials}`,
            },
            body: JSON.stringify({
                email: userEmail,
                enabled: !notificationsEnabled
            })
        })
            .then(res => res.json())
            .then(data =>
            {
                setNotificationsEnabled(!notificationsEnabled);
                setLoading(false);
            })
            .catch(err =>
            {
                console.error('Failed to update notification preferences:', err);
                setLoading(false);
            });
    };

    return (
        <div className={`modal ${isOpen ? 'is-active' : ''}`}>
            <div className="modal-background" onClick={onClose}></div>
            <div className="modal-card" style={{ maxWidth: '500px' }}>
                <header className="modal-card-head" style={{ background: '#f8f9fa' }}>
                    <p className="modal-card-title">
                        <span className="icon-text">
                            <span className="icon">
                                <i className="fas fa-user-cog"></i>
                            </span>
                            <span>User Settings</span>
                        </span>
                    </p>
                    <button
                        className="delete"
                        aria-label="close"
                        onClick={onClose}
                    ></button>
                </header>

                <section className="modal-card-body">
                    <div className="box">
                        <h3 className="title is-5 mb-3">Email Notifications</h3>
                        <div className="field">
                            <label className="checkbox">
                                <input
                                    type="checkbox"
                                    checked={notificationsEnabled}
                                    onChange={handleToggleNotifications}
                                    disabled={loading}
                                    className="mr-2"
                                />
                                Receive email notifications when you are bumped from a room
                                <p className="help">You will receive an email when someone bumps you from your current room</p>
                            </label>
                        </div>
                    </div>
                </section>

                <footer className="modal-card-foot" style={{ background: '#f8f9fa', justifyContent: 'flex-end' }}>
                    <button
                        className="button is-primary"
                        onClick={onClose}
                    >
                        <span className="icon">
                            <i className="fas fa-check"></i>
                        </span>
                        <span>Close</span>
                    </button>
                </footer>
            </div>
        </div>
    );
};

export default UserSettingsModal; 