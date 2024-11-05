package server

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRes struct {
	UserID int    `json:"userId"`
	Token  string `json:"token"`
}

type SignUpReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Mobile   string `json:"mobile"`
	Password string `json:"password"`
}

type SignUpRes struct {
	Msg    string `json:"msg"`
	Token  string `json:"token"`
	UserID int    `json:"userId"`
}

type UpdatePasswordReq struct {
	Email       string `json:"email"`
	NewPassword string `json:"newPassword"`
}

type AddTicketReq struct {
	EventDate       string `json:"eventDate"`
	Venue           string `json:"venue"`
	NumberOfTickets int    `json:"numberOfTickets"`
	Price           int    `json:"price"`
	SeatInfo        []Seat `json:"seatInfo"`
	Deadline        string `json:"deadline"`
}

type Seat struct {
	SeatNumber int    `json:"seatNumber"`
	Block      string `json:"block"`
	Level      int    `json:"level"`
}

type AddTicketRes struct {
	TicketId int `json:"ticketId"`
}

type GetAllTicketsReq struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type GetAllTicketsRes struct {
	Tickets []*Ticket
}

type Ticket struct {
	TicketID        int    `json:"ticketId"`
	EventDate       string `json:"eventDate"`
	UserID          int    `json:"userId"`
	Venue           string `json:"venue"`
	NumberOfTickets int    `json:"numberOfTickets"`
	Price           int    `json:"price"`
	HighestBid      int    `json:"highestBid"`
	SeatInfo        []Seat `json:"seatInfo"`
	Deadline        string `json:"deadline"`
	ListedBy        string `json:"listedBy"`
}

type AddNewBidReq struct {
	TicketID      int    `json:"ticketId"`
	OwnerID       int    `json:"userId"`
	BidPrice      int    `json:"bidPrice"`
	OriginalPrice int    `json:"originalPrice"`
	Venue         string `json:"venue"`
}
type AddNewBidRes struct {
	BidID int `json:"bidId"`
}

type GetUserBidsRes struct {
	Bids []UserBid `json:"bids"`
}
type UserBid struct {
	BidId         int    `json:"BidId"`
	TicketId      int    `json:"ticketId"`
	BidPrice      int    `json:"bidPrice"`
	OriginalPrice int    `json:"originalPrice"`
	Venue         string `json:"venue"`
	OwnerId       int    `json:"ownerId"`
	BidderId      int    `json:"bidderId"`
	CreatedAt     string `json:"createdAt"`
}
