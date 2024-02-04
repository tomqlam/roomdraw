import logo from './logo.svg';
import React from 'react';
import { useState, useContext } from 'react';
import { MyContext } from './MyContext';
import 'bulma/css/bulma.min.css';
import BumpModal from './BumpModal';

function FloorGrid({ gridData }) {
  // Define your data structure with columns
  const {
    setIsModalOpen,
    setSelectedItem,
    selectedOccupants,
    setSelectedOccupants,
    setSelectedSuiteObject,
    getNameById,
    setSelectedRoomObject,
    setPullMethod,
    cellColors,
    selectedID,
    setSelectedID,
    onlyShowBumpableRooms,
    userMap
  } = useContext(MyContext);


  function getOccupantsByRoomNumber(roomNumber) {
    // Iterate over each suite
    for (let suite of gridData.suites) {
      // Find the room with the given room number within the current suite
      console.log("FIRST ITERATION");
      console.log(suite);
      const room = suite.rooms.find(r => r.roomNumber.toString() === roomNumber.toString());

      // If the room exists, return the list of occupants
      if (room) {
        setSelectedRoomObject(room);
        setSelectedSuiteObject(suite);
        return [room.occupant1.toString(), room.occupant2.toString(), room.occupant3.toString(), room.occupant4.toString()];
      }
    }
    console.log("did not find the occupants");

    // If the room does not exist in any suite, return an empty array
    return ['', '', '', ''];
  }

  const gridContainerStyle = {
    display: 'grid',
    gridTemplateColumns: 'auto auto auto auto auto auto auto', // Adjust the number of 'auto' as needed
    gap: '5px',
    maxWidth: '800px', // Set the maximum width of the grid container
    margin: '0 auto', // Center the grid container horizontally

  };

  const gridItemStyle = {
    //border: '1px solid #ddd',
    borderRadius: '3px',
    padding: '2px',
    backgroundColor: cellColors.occupants, // Set the background color of the cells
    textAlign: 'center',
    fontSize: '15px', // Set the font size of the cells
    color: '#000000',

  };

  function darken(color, factor) {
    const f = parseInt(factor, 10) || 0;
    const RGB = color.substring(1).match(/.{2}/g);
    const newColor = RGB.map((c) => {
      const hex = Math.max(0, Math.min(255, parseInt(c, 16) - f)).toString(16);
      return hex.length === 1 ? `0${hex}` : hex;
    });
    return `#${newColor.join('')}`;
  }
  
  const getGridItemStyle = (occupancy, maxOccupancy, suiteIndex, pullPriority) => {
    if (occupancy < maxOccupancy){
      return {
        ...gridItemStyle,
        backgroundColor: cellColors.unbumpableRoom
      };
    }
    let backgroundColor = (suiteIndex % 2 === 0 ? cellColors.evenSuite : cellColors.oddSuite);
    
    if (!checkBumpable(pullPriority) && onlyShowBumpableRooms) {
      backgroundColor = darken(backgroundColor, 100); // darken the color by 10%
    }
  
    return {
      ...gridItemStyle,
      backgroundColor
    };
  };

  const roomNumberStyle = {
    ...gridItemStyle,
    backgroundColor: cellColors.roomNumber, // Change the color to your desired color
  };
  const pullMethodStyle = {
    ...gridItemStyle,
    backgroundColor: cellColors.pullMethod, // Change the color to your desired color
  };
  const handleCellClick = (roomNumber) => {
    setIsModalOpen(true);
    setSelectedItem(roomNumber);
    setSelectedOccupants(getOccupantsByRoomNumber(roomNumber));
    console.log(selectedOccupants);
    console.log('lol');
    setPullMethod("Select a pull method");

  };

  function getPullMethodByRoomNumber(roomNumber) {
    // TODO FINISH 
    // Iterate over each suite
    for (let suite of gridData.suites) {
      // Find the room with the given room number within the current suite
      const room = suite.rooms.find(r => r.roomNumber.toString() === roomNumber.toString());

      // If the room exists, return the list of occupants
      

      if (room) {
        var pullPriority = room.pullPriority;
        var finalString = "";
        if (pullPriority.pullType === 2){
          pullPriority = pullPriority.inherited;
        }
        if (pullPriority.isPreplaced) {
          finalString += "Preplaced";
        }
        if (pullPriority.hasInDorm) {
          finalString += `In-Dorm ${pullPriority.drawNumber}`;
          
        } else {
          const yearMapping = ["", "", "Sophomore", "Junior", "Senior"];
        finalString += `${yearMapping[pullPriority.year]} ${pullPriority.drawNumber !== 0 ? pullPriority.drawNumber : ''}`;

        }
        
        return finalString += `${room.pullPriority.pullType === 2 ? " Pull" : ''}`;
      }
    }

    // If the room does not exist in any suite, return an empty array
    return 'n/a';
  }

  const checkBumpable = (pullPriority) => {
    if (!pullPriority.valid)  {
      return true;
    }
    if (pullPriority.isPreplaced) {
      return false;
    }
    if (pullPriority.hasInDorm) {
      if (!userMap[selectedID].InDorm) {
        return false;
      }
    }
    // just compare the numbers
    return userMap[selectedID].DrawNumber <= pullPriority.drawNumber;

  }



  return (
    <div style={gridContainerStyle}>
      {/* <div style={gridItemStyle}><strong></strong></div>

        {/* begin filler code that does nothing*/}
      {/* <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div>
        <div style={gridItemStyle}><strong>{}</strong></div> */}




      <div style={roomNumberStyle}><strong>Room Number</strong></div>
      <div style={roomNumberStyle}><strong>Pull Method</strong></div>
      <div style={roomNumberStyle}><strong>Suite</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 1</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 2</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 3</strong></div>
      <div style={roomNumberStyle}><strong>Occupant 4</strong></div>

      {gridData.suites.map((suite, suiteIndex) => (
        suite.rooms.sort((a, b) => Number(a.roomNumber) - Number(b.roomNumber))  // Sort the rooms by room number
        .map((room, roomIndex) => (
          <React.Fragment key={roomIndex}>
            <div
              style={getGridItemStyle(room.maxOccupancy, 1, suiteIndex, room.pullPriority)}
              onClick={() => handleCellClick(room.roomNumber)}
            >
              {room.roomNumber}
            </div>
            <div style={getGridItemStyle(room.maxOccupancy, 1, suiteIndex, room.pullPriority)} onClick={() => handleCellClick(room.roomNumber)}>{getPullMethodByRoomNumber(room.roomNumber)}</div>
            {
              roomIndex === 0
              && <div style={{
                ...pullMethodStyle, gridRow: `span ${suite.rooms.length}`, backgroundColor: suiteIndex % 2 === 0
                  ? cellColors.evenSuite // color for even suiteIndex
                  : cellColors.oddSuite
              }} >Insert suite name</div>

            }
            <div style={getGridItemStyle(room.maxOccupancy, 1, suiteIndex, room.pullPriority)} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant1)}</div>
            <div style={getGridItemStyle(room.maxOccupancy, 2, suiteIndex, room.pullPriority)} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant2)}</div>
            <div style={getGridItemStyle(room.maxOccupancy, 3, suiteIndex, room.pullPriority)} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant3)}</div>
            <div style={getGridItemStyle(room.maxOccupancy, 4, suiteIndex, room.pullPriority)} onClick={() => handleCellClick(room.roomNumber)}>{getNameById(room.occupant4)}</div>

          </React.Fragment>
        ))
      ))}





      {/* {gridData.map((item, index) => (
        <React.Fragment key={index}>
          <div style={roomNumberStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.roomNumber}</div>
          <div style={pullMethodStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.notes}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant1}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant2}</div>
          <div style={gridItemStyle} onClick={() => handleCellClick(item.roomNumber)}>{item.occupant3}</div>
        </React.Fragment>
      ))} */}
    </div>

  );
}

export default FloorGrid;
