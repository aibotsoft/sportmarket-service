create or alter proc dbo.uspSaveBet @SurebetId bigint,
                                    @SideIndex tinyint,

                                    @BetId bigint,
                                    @TryCount tinyint= null,
                                    @Status varchar(1000)= null,
                                    @StatusInfo varchar(1000)= null,
                                    @Start bigint= null,
                                    @Done bigint= null,
                                    @Price decimal(9, 5)= null,
                                    @Stake decimal(9, 5)= null,
                                    @ApiBetId bigint= null
as
begin
    set nocount on
    MERGE dbo.Bet AS t
    USING (select @SurebetId, @SideIndex) s (SurebetId, SideIndex)
    ON (t.SurebetId = s.SurebetId and t.SideIndex = s.SideIndex)

    WHEN MATCHED THEN
        UPDATE
        SET BetId = @BetId,
            TryCount   = @TryCount,
            Status     = @Status,
            StatusInfo = @StatusInfo,
            Start      = @Start,
            Done       = @Done,
            Price      = @Price,
            Stake      = @Stake,
            ApiBetId   = @ApiBetId,
            UpdatedAt  =sysdatetimeoffset()

    WHEN NOT MATCHED THEN
        INSERT (SurebetId, SideIndex, BetId, TryCount, Status, StatusInfo, Start, Done, Price, Stake, ApiBetId)
        VALUES (s.SurebetId, @SideIndex, @BetId, @TryCount, @Status, @StatusInfo, @Start, @Done, @Price, @Stake, @ApiBetId);
end