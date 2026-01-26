import { jwtDecode } from 'jwt-decode';
import { useContext, useState } from 'react';
import { MyContext } from '../context/MyContext';


const AdminBumpModalFunctions = ({ closeModal }) =>
{
    const {
        adminList,
        credentials,

        selectedRoomObject,
        selectedOccupants,
        setRefreshKey,

        handleErrorFromTokenExpiry,
        setPullError,
        setShowModalError,


    } = useContext(MyContext);

    // Add local loading state
    const [loading, setLoading] = useState({
        frosh: false,
        preplace: false,
        removePreplace: false
    });

    function postToFrosh(roomObject)
    {
        setLoading(prev => ({ ...prev, frosh: true }));
        fetch(`${process.env.REACT_APP_API_URL}/frosh/${roomObject.roomUUID}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
            },
        })
            .then(response => response.json())
            .then(data =>
            {
                setLoading(prev => ({ ...prev, frosh: false }));
                if (data.error)
                {
                    // Handle error case
                    setPullError(data.error);
                    setShowModalError(true);
                    return;
                }
                closeModal();
                setRefreshKey(prev => prev + 1);
                if (handleErrorFromTokenExpiry(data))
                {
                    return;
                };
            })
            .catch((error) =>
            {
                console.error('Error:', error);
                setLoading(prev => ({ ...prev, frosh: false }));
                setPullError("An unexpected error occurred. Please try again.");
                setShowModalError(true);
            });
    }

    function preplaceOccupants(roomObject)
    {
        setLoading(prev => ({ ...prev, preplace: true }));
        fetch(`${process.env.REACT_APP_API_URL}/rooms/preplace/${roomObject.roomUUID}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
            },
            body: JSON.stringify({
                proposedOccupants: selectedOccupants
                    .filter(occupant => occupant !== '')
                    .map(occupant => Number(occupant.value)),
            }),
        })
            .then(response => response.json())
            .then(data =>
            {
                setLoading(prev => ({ ...prev, preplace: false }));
                if (data.error)
                {
                    // Handle error case
                    setPullError(data.error);
                    setShowModalError(true);
                    return;
                }
                closeModal();
                setRefreshKey(prev => prev + 1);
                if (handleErrorFromTokenExpiry(data))
                {
                    return;
                };
            })
            .catch((error) =>
            {
                console.error('Error:', error);
                setLoading(prev => ({ ...prev, preplace: false }));
                setPullError("An unexpected error occurred. Please try again.");
                setShowModalError(true);
            });
    }

    function removePreplaceOccupants(roomObject)
    {
        setLoading(prev => ({ ...prev, removePreplace: true }));
        fetch(`${process.env.REACT_APP_API_URL}/rooms/preplace/remove/${roomObject.roomUUID}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
            }
        })
            .then(response => response.json())
            .then(data =>
            {
                setLoading(prev => ({ ...prev, removePreplace: false }));
                if (data.error)
                {
                    // Handle error case
                    setPullError(data.error);
                    setShowModalError(true);
                    return;
                }
                closeModal();
                setRefreshKey(prev => prev + 1);
                if (handleErrorFromTokenExpiry(data))
                {
                    return;
                };
            })
            .catch((error) =>
            {
                console.error('Error:', error);
                setLoading(prev => ({ ...prev, removePreplace: false }));
                setPullError("An unexpected error occurred. Please try again.");
                setShowModalError(true);
            });
    }

    const isAdmin = adminList.includes(jwtDecode(credentials).email);

    // Only render these buttons if the user is an admin

    if (!isAdmin)
    {
        return null;
    }

    return (
        <>
            <label className="label">Admin-Only Functions</label>

            <div className="buttons">
                <button className={`button is-warning ${loading.frosh ? 'is-loading' : ''}`} style={{ marginBottom: '0.5rem' }} onClick={() => postToFrosh(selectedRoomObject)}>Add Frosh</button>
                <button className={`button is-warning ${loading.preplace ? 'is-loading' : ''}`} style={{ marginBottom: '0.5rem' }} onClick={() => preplaceOccupants(selectedRoomObject)}>Pre-Place Occupants</button>
                <button className={`button is-warning ${loading.removePreplace ? 'is-loading' : ''}`} style={{ marginBottom: '0.5rem' }} onClick={() => removePreplaceOccupants(selectedRoomObject)}>Remove Pre-Placed Occupants</button>
            </div>
            <p className="help is-danger">These are dangerous: be sure before toggling!</p>

        </>
    );
};

export default AdminBumpModalFunctions;