$(document).ready(function() {    
   var cy = cytoscape({

    container: document.getElementById('cy'), // container to render in

    elements: [ // list of graph elements to start with
      { // node a
        data: { id: 'a' }
      },
      { // node b
        data: { id: 'b' }
      },
      { // edge ab
        data: { id: 'ab', source: 'a', target: 'b' }
      }
    ],

    style: [ // the stylesheet for the graph
      {
        selector: 'node',
        style: {
          'background-color': '#666',
          'width': '100%',
          'label': 'data(id)'
        }
      },

      {
        selector: 'edge',
        style: {
          'width': 100,
          'line-color': '#ccc',
          'target-arrow-color': '#ccc',
          'target-arrow-shape': 'triangle',
          'curve-style': 'bezier'
        }
      }
    ],

    layout: {
      name: 'grid',
      rows: 1
    }

  });
  cy.add({
    group: "nodes",
    data: { id: 'x' },
    position: { x: 190 , y: 190  }
});
  cy.resize()
})