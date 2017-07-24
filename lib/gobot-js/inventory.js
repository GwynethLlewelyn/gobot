// specify the columns
var columnDefs = [
      {headerName:'UUID', field:'UUID', width:80, editable:true, sortable:true}, 
      {headerName:'Name', field:'Name', width:200, editable:true, sortable:true},
      {headerName:'Type', field:'Type', width:100, editable:true},
      {headerName:'Permissions', field:'Permissions', width:100, editable:true},
      {headerName:'LastUpdate', field:'LastUpdate', width:100, editable:true, sortable:true, sort: 'desc'}
];

var URLPathPrefix;

var gridOptions = {
	columnDefs: columnDefs,
	rowData: null, // change to null since we're filling it from our own servers
	rowSelection: 'multiple',
	enableColResize: true,
	enableSorting: true,
	enableFilter: true,
	enableRangeSelection: true,
	rowHeight: 22,
	animateRows: true,
	debug: true,
		editType: 'fullRow',
    onRowValueChanged: function(event) {
        var data = event.data;
        console.log('onRowValueChanged: (' + data.UUID + ', ' + data.Name + ', ' + data.Type + ' ...)');
        
        var httpRequest = new XMLHttpRequest(); // see https://stackoverflow.com/questions/6418220/javascript-send-json-object-with-ajax
		httpRequest.open('POST', URLPathPrefix + '/uiInventoryUpdate/');
		httpRequest.setRequestHeader("Content-Type", "application/json");
		httpRequest.onreadystatechange = function() { // see https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest/onreadystatechange
			if (httpRequest.readyState === XMLHttpRequest.DONE && httpRequest.status === 200) {
				console.log(httpRequest.responseText);
			}
		};
		var response = JSON.stringify(data);
		// console.log('Response is going to be: ' + response)
		httpRequest.send(response);
    },
	onGridReady: function() {
		gridOptions.api.sizeColumnsToFit();
	}
};

// wait for the document to be loaded, otherwise ag-Grid will not find the div in the document.
document.addEventListener("DOMContentLoaded", function() {
	URLPathPrefix = document.getElementById("URLPathPrefix").innerText;

	// lookup the container we want the Grid to use
	var eGridDiv = document.querySelector('#inventoryGrid');

	// create the grid passing in the div to use together with the columns & data we want to use
	new agGrid.Grid(eGridDiv, gridOptions);
	
	// do http request to get our sample data (see https://www.ag-grid.com/javascript-grid-value-getters/?framework=all#gsc.tab=0)
	var httpRequest = new XMLHttpRequest();
	httpRequest.open('GET', URLPathPrefix + '/uiInventory/');
	httpRequest.send();
	httpRequest.onreadystatechange = function() {
		if (httpRequest.readyState == 4 && httpRequest.status == 200) {
			var httpResult = JSON.parse(httpRequest.responseText);
			gridOptions.api.setRowData(httpResult);
		}
	};
});

function onRemoveSelected() {
	BootstrapDialog.confirm({
			title: 'WARNING',
			message: 'Are you sure you want to delete the selected rows?',
			type: BootstrapDialog.TYPE_WARNING,
			size: BootstrapDialog.SIZE_SMALL,
			closable: true,
			draggable: true,
			btnCancelLabel: 'Cancel',
			btnCancelIcon: 'glyphicon glyphicon-remove',
			btnOKLabel: 'Ok',
			btnOkIcon: 'glyphicon glyphicon-ok',
			btnOKClass: 'btn-warning',
			callback: function(result) {
								if (result) {
					var selectedRows = gridOptions.api.getSelectedRows();
					var ids = "";
					
					selectedRows.forEach( function(selectedRow, index) {
						if (index!==0) {
							ids += '", "';
						}
						else {
							ids = '"';
						}
						ids += selectedRow.UUID;
					});
					ids += '"';

					// console.log('Removing UUIDs: ' + ids);
					
					var httpRequest = new XMLHttpRequest();
					httpRequest.open('POST', URLPathPrefix + '/uiInventoryRemove/');
					httpRequest.setRequestHeader("Content-Type", "text/plain");
					httpRequest.onreadystatechange = function() {
						if (httpRequest.readyState === XMLHttpRequest.DONE) {
							if (httpRequest.status === 200) {
								console.log(httpRequest.responseText);
								gridOptions.api.updateRowData({remove: selectedRows});
								gridOptions.api.refreshView();
							} else {
								console.log(httpRequest.responseText);
								BootstrapDialog.show({
									title: httpRequest.status,
									message: 'Error: ' + httpRequest.responseText,
									type: BootstrapDialog.TYPE_DANGER
								});
							}
						}
					};
					httpRequest.send(ids);
					
					gridOptions.api.updateRowData({remove: selectedRows});
					gridOptions.api.refreshView();
				} else {
					console.log("Canceled removing nodes.");
				}
			}
		});
}

function onInsertRow() {
	var dateNow = new Date();
	var newItem = {
		UUID: "00000000-0000-0000-0000-000000000000",
		Name: "empty",
		LastUpdate: dateNow.toISOString()
	};
	gridOptions.api.updateRowData({add: [newItem], addIndex: 0});
	gridOptions.api.setFocusedCell(0, newItem, 'top');
	var newItemNode = gridOptions.api.getDisplayedRowAtIndex(0);
	newItemNode.setSelected(true);
	gridOptions.api.ensureIndexVisible(0);
	gridOptions.api.refreshView();
}