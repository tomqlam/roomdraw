import {jwtDecode} from 'jwt-decode';
import { useContext } from 'react';
import { MyContext } from './MyContext';


const AdminBumpModalFunctions = ({closeModal}) => {
    const {
        adminList,
        credentials,

        selectedRoomObject,
        selectedOccupants,
        setRefreshKey,
  
        handleErrorFromTokenExpiry,


    } = useContext(MyContext);

    function postToFrosh(roomObject) {
        fetch(`/frosh/${roomObject.roomUUID}`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
          },
        })
          .then(response => response.json())
          .then(data => {
            // commented console.log (data);
            closeModal();
            setRefreshKey(prev => prev + 1);
            if (handleErrorFromTokenExpiry(data)) {
              return;
            };
          })
          .catch((error) => {
            console.error('Error:', error);
          });
      }
    
      function preplaceOccupants(roomObject) {
        fetch(`/rooms/preplace/${roomObject.roomUUID}`, {
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
          .then(data => {
            // commented console.log (data);
            closeModal();
            setRefreshKey(prev => prev + 1);
            if (handleErrorFromTokenExpiry(data)) {
              return;
            };
          })
          .catch((error) => {
            console.error('Error:', error);
          });
      }
    
      function removePreplaceOccupants(roomObject) {
        fetch(`/rooms/preplace/${roomObject.roomUUID}`, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('jwt')}`,
          },
          body: JSON.stringify({
            proposedOccupants: [],
          }),
        })
          .then(response => response.json())
          .then(data => {
            // commented console.log (data);
            closeModal();
            setRefreshKey(prev => prev + 1);
            if (handleErrorFromTokenExpiry(data)) {
              return;
            };
          })
          .catch((error) => {
            console.error('Error:', error);
          });
      }

    const isAdmin = adminList.includes(jwtDecode(credentials).email);

    // Only render these buttons if the user is an admin

    if (!isAdmin) {
        return null;
    }

    return (
        <>
            <label className="label">Admin-Only Functions</label>

            <button className="button is-warning" style={{ marginRight: '5px' }} onClick={() => postToFrosh(selectedRoomObject)}>Add Frosh</button>
            <button className="button is-warning" style={{ marginRight: '5px' }} onClick={() => preplaceOccupants(selectedRoomObject)}>Pre-Place Occupants</button>
            <button className="button is-warning" onClick={() => removePreplaceOccupants(selectedRoomObject)}>Remove Pre-Placed Occupants</button>
            <p class="help is-danger">These are dangerous: be sure before toggling!</p>

        </>
    );
};

export default AdminBumpModalFunctions;